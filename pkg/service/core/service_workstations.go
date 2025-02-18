package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/api/googleapi"

	"github.com/navikt/nada-backend/pkg/normalize"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

const (
	maxAttemptsToGetVMs = 12
)

var _ service.WorkstationsService = (*workstationService)(nil)

type workstationService struct {
	workstationsProject          string
	workstationsLoggingBucket    string
	workstationsLoggingView      string
	serviceAccountsProject       string
	location                     string
	tlsSecureWebProxyPolicy      string
	firewallPolicyName           string
	administratorServiceAcccount string
	artifactRepositoryName       string
	artifactRepositoryProject    string
	signerServiceAccount         string
	podName                      string

	workstationsQueue       service.WorkstationsQueue
	workstationStorage      service.WorkstationsStorage
	workstationAPI          service.WorkstationsAPI
	serviceAccountAPI       service.ServiceAccountAPI
	secureWebProxyAPI       service.SecureWebProxyAPI
	cloudResourceManagerAPI service.CloudResourceManagerAPI
	computeAPI              service.ComputeAPI
	cloudLoggingAPI         service.CloudLoggingAPI
	artifactRegistryAPI     service.ArtifactRegistryAPI
	datavarehusAPI          service.DatavarehusAPI
	iamcredentialsAPI       service.IAMCredentialsAPI
	cloudBillingAPI         service.CloudBillingAPI
}

func (s *workstationService) GetWorkstationURLList(ctx context.Context, user *service.User) (*service.WorkstationURLList, error) {
	const op errs.Op = "workstationService.GetWorkstationURLList"

	output, err := s.workstationStorage.GetLastWorkstationsURLList(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return output, nil
}

func (s *workstationService) GetWorkstationOnpremMapping(ctx context.Context, user *service.User) (*service.WorkstationOnpremAllowList, error) {
	const op errs.Op = "workstationService.GetWorkstationOnpremMapping"

	hosts, err := s.workstationStorage.GetLastWorkstationsOnpremAllowList(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationOnpremAllowList{
		Hosts: hosts,
	}, nil
}

func (s *workstationService) UpdateWorkstationOnpremMapping(ctx context.Context, user *service.User, onpremAllowList *service.WorkstationOnpremAllowList) error {
	const op errs.Op = "workstationService.UpdateWorkstationOnpremMapping"

	slug := user.Ident

	err := s.workstationStorage.CreateWorkstationsOnpremAllowListChange(ctx, slug, onpremAllowList.Hosts)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *workstationService) GetWorkstationStartJob(ctx context.Context, id int64) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationService.GetWorkstationStartJob"

	job, err := s.workstationsQueue.GetWorkstationStartJob(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (s *workstationService) CreateWorkstationStartJob(ctx context.Context, user *service.User) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationService.CreateWorkstationStartJob"

	_, err := s.GetWorkstation(ctx, user)
	if err != nil {
		return nil, err
	}

	job, err := s.workstationsQueue.CreateWorkstationStartJob(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (s *workstationService) GetWorkstationStartJobsForUser(ctx context.Context, ident string) (*service.WorkstationStartJobs, error) {
	const op errs.Op = "workstationService.GetWorkstationStartJobsForUser"

	jobs, err := s.workstationsQueue.GetWorkstationStartJobsForUser(ctx, ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationStartJobs{
		Jobs: jobs,
	}, nil
}

func (s *workstationService) GetWorkstationJobsForUser(ctx context.Context, ident string) (*service.WorkstationJobs, error) {
	const op errs.Op = "workstationService.GetRunningWorkstationJobsForUser"

	jobs, err := s.workstationsQueue.GetWorkstationJobsForUser(ctx, ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationJobs{
		Jobs: jobs,
	}, nil
}

func (s *workstationService) CreateWorkstationJob(ctx context.Context, user *service.User, input *service.WorkstationInput) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationService.CreateWorkstationJob"

	job, err := s.workstationsQueue.CreateWorkstationJob(ctx, &service.WorkstationJobOpts{
		User:  user,
		Input: input,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (s *workstationService) GetWorkstationJob(ctx context.Context, jobID int64) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationService.GetWorkstationJob"

	job, err := s.workstationsQueue.GetWorkstationJob(ctx, jobID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (s *workstationService) GetWorkstationLogs(ctx context.Context, user *service.User) (*service.WorkstationLogs, error) {
	const op errs.Op = "workstationService.GetWorkstationLogs"

	slug := user.Ident

	oneHourAgo := time.Now().Add(-time.Hour).Format(time.RFC3339)

	raw, err := s.cloudLoggingAPI.ListLogEntries(ctx, s.workstationsProject, &service.ListLogEntriesOpts{
		ResourceNames: []string{
			service.WorkstationDeniedRequestsLoggingResourceName(s.workstationsProject, s.location, s.workstationsLoggingBucket, s.workstationsLoggingView),
		},
		Filter: service.WorkstationDeniedRequestsLoggingFilter(s.tlsSecureWebProxyPolicy, slug, oneHourAgo),
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	uniqueHostPaths := map[string]*service.LogEntry{}

	for _, entry := range raw {
		if entry.HTTPRequest == nil {
			continue
		}

		hostPath, err := url.JoinPath(entry.HTTPRequest.URL.Host, entry.HTTPRequest.URL.Path)
		if err != nil {
			return nil, errs.E(op, err)
		}

		uniqueHostPaths[hostPath] = entry
	}

	hosts := make([]*service.LogEntry, 0, len(uniqueHostPaths))
	for _, logEntry := range uniqueHostPaths {
		hosts = append(hosts, logEntry)
	}

	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].Timestamp.After(hosts[j].Timestamp)
	})

	return &service.WorkstationLogs{
		ProxyDeniedHostPaths: hosts,
	}, nil
}

func (s *workstationService) GetWorkstationOptions(ctx context.Context) (*service.WorkstationOptions, error) {
	const op errs.Op = "workstationService.GetWorkstationOptions"

	images, err := s.artifactRegistryAPI.ListContainerImagesWithTag(ctx, &service.ContainerRepositoryIdentifier{
		Project:    s.artifactRepositoryProject,
		Location:   s.location,
		Repository: s.artifactRepositoryName,
	}, service.WorkstationImagesTag)
	if err != nil {
		return nil, errs.E(op, err)
	}

	containerImages := make([]*service.WorkstationContainer, len(images))
	for i, image := range images {
		containerImages[i] = &service.WorkstationContainer{
			Image:         image.URI,
			Description:   image.Name,
			Labels:        image.Manifest.Labels,
			Documentation: image.Documentation,
		}
	}

	globalURLAllowList, err := s.secureWebProxyAPI.GetURLList(ctx, &service.URLListIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Slug:     service.GlobalURLAllowListName,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	hourlyCost, err := s.cloudBillingAPI.GetHourlyCostInNOKFromSKU(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationOptions{
		MachineTypes:       service.WorkstationMachineTypes(hourlyCost.CostForConfiguration),
		ContainerImages:    containerImages,
		GlobalURLAllowList: globalURLAllowList,
	}, nil
}

func (s *workstationService) StartWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.StartWorkstation"

	slug := user.Ident

	err := s.workstationAPI.StartWorkstation(ctx, &service.WorkstationIdentifier{
		Slug:                  slug,
		WorkstationConfigSlug: slug,
	})
	if err != nil {
		return errs.E(op, err)
	}

	config, err := s.workstationAPI.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("getting workstation config: %w", err))
	}

	var vms []*service.VirtualMachine

	for attempts := 0; ; attempts++ {
		vms, err = s.computeAPI.GetVirtualMachinesByLabel(ctx, config.ReplicaZones, &service.Label{
			Key:   service.WorkstationConfigIDLabel,
			Value: slug,
		})
		if err != nil {
			return errs.E(op, fmt.Errorf("getting virtual machines by label: %w", err))
		}

		if len(vms) == 1 {
			break
		}

		if attempts > maxAttemptsToGetVMs {
			return errs.E(op, fmt.Errorf("for user %s expected exactly one VM, got %d", slug, len(vms)))
		}

		time.Sleep(10 * time.Second)
	}

	err = s.workstationStorage.CreateWorkstationsActivity(ctx, slug, strconv.FormatUint(vms[0].ID, 10), service.WorkstationActionTypeStart)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *workstationService) GetWorkstationZonalTagBindingsJobsForUser(ctx context.Context, ident string) ([]*service.WorkstationZonalTagBindingsJob, error) {
	const op errs.Op = "workstationService.GetWorkstationZonalTagBindingsJobsForUser"

	jobs, err := s.workstationsQueue.GetWorkstationZonalTagBindingsJobsForUser(ctx, ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return jobs, nil
}

func (s *workstationService) GetWorkstationZonalTagBindings(ctx context.Context, ident string) ([]*service.EffectiveTag, error) {
	const op errs.Op = "workstationService.GetWorkstationZonalTagBindings"

	slug := ident

	config, err := s.workstationAPI.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: slug,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	vms, err := s.computeAPI.GetVirtualMachinesByLabel(ctx, config.ReplicaZones, &service.Label{
		Key:   service.WorkstationConfigIDLabel,
		Value: slug,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	currentTagNamespacedName := map[string][]*service.EffectiveTag{}
	for _, vm := range vms {
		tags, err := s.cloudResourceManagerAPI.ListEffectiveTags(ctx, vm.Zone, fmt.Sprintf("//compute.googleapis.com/projects/%s/zones/%s/instances/%d", s.workstationsProject, vm.Zone, vm.ID))
		if err != nil {
			return nil, errs.E(op, err)
		}

		for _, tag := range tags {
			if tag.TagKeyParentName == service.WorkstationEffectiveTagGCPKeyParentName {
				// We are only interested in the tags we have added, and these are managed by GCP, so we skip them
				continue
			}

			currentTagNamespacedName[vm.Name] = append(currentTagNamespacedName[vm.Name], tag)
		}
	}

	var tags []*service.EffectiveTag

	for _, vm := range vms {
		tags = append(tags, currentTagNamespacedName[vm.Name]...)
	}

	return tags, nil
}

func (s *workstationService) CreateWorkstationZonalTagBindingsJobForUser(ctx context.Context, ident, requestID string, input *service.WorkstationOnpremAllowList) (*service.WorkstationZonalTagBindingsJob, error) {
	const op errs.Op = "workstationService.CreateWorkstationZonalTagBindingsJobForUser"

	job, err := s.workstationsQueue.CreateWorkstationZonalTagBindingsJob(ctx, ident, requestID, input.Hosts)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (s *workstationService) UpdateWorkstationZonalTagBindingsForUser(ctx context.Context, ident string, requestID string, hosts []string) error {
	const op errs.Op = "workstationService.CreateWorkstationZonalTagBindingJobsForUser"

	slug := ident

	// We want these tags to be present on the VMs
	expectedTagNamespacedName := map[string]struct{}{}
	for _, tag := range hosts {
		expectedTagNamespacedName[fmt.Sprintf("%s/%s/%s", s.workstationsProject, tag, tag)] = struct{}{}
	}

	config, err := s.workstationAPI.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("getting workstation config: %w", err))
	}

	vms, err := s.computeAPI.GetVirtualMachinesByLabel(ctx, config.ReplicaZones, &service.Label{
		Key:   service.WorkstationConfigIDLabel,
		Value: slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("getting virtual machines by label: %w", err))
	}

	// We want to know the current tags on the VMs
	currentTagNamespacedName := map[string][]*service.EffectiveTag{}
	for _, vm := range vms {
		tags, err := s.cloudResourceManagerAPI.ListEffectiveTags(ctx, vm.Zone, fmt.Sprintf("//compute.googleapis.com/projects/%s/zones/%s/instances/%d", s.workstationsProject, vm.Zone, vm.ID))
		if err != nil {
			return errs.E(op, fmt.Errorf("listing effective tags: %w", err))
		}

		for _, tag := range tags {
			if tag.TagKeyParentName == service.WorkstationEffectiveTagGCPKeyParentName {
				// We are only interested in the tags we have added, and these are managed by GCP, so we skip them
				continue
			}

			currentTagNamespacedName[vm.Name] = append(currentTagNamespacedName[vm.Name], tag)
		}
	}

	if len(vms) != 1 {
		return errs.E(op, fmt.Errorf("for user %s expected exactly one VM, got %d", ident, len(vms)))
	}

	vm := vms[0]

	// Add tags that are missing
	for tag := range expectedTagNamespacedName {
		found := false
		for _, currentTag := range currentTagNamespacedName[vm.Name] {
			if currentTag.NamespacedTagValue == tag {
				found = true

				break
			}
		}

		if !found {
			err := s.AddWorkstationZonalTagBinding(ctx,
				vm.Zone,
				fmt.Sprintf("//compute.googleapis.com/projects/%s/zones/%s/instances/%d", s.workstationsProject, vm.Zone, vm.ID),
				tag,
			)
			if err != nil {
				return fmt.Errorf("adding workstation zonal tag binding: %w", err)
			}
		}
	}

	// Remove tags that are not supposed to be there
	for _, currentTag := range currentTagNamespacedName[vm.Name] {
		if _, hasKey := expectedTagNamespacedName[currentTag.NamespacedTagValue]; !hasKey {
			err := s.RemoveWorkstationZonalTagBinding(ctx,
				vm.Zone,
				fmt.Sprintf("//compute.googleapis.com/projects/%s/zones/%s/instances/%d", s.workstationsProject, vm.Zone, vm.ID),
				currentTag.TagValue,
			)
			if err != nil {
				return errs.E(op, fmt.Errorf("removing zonal tag binding: %w", err))
			}
		}
	}

	tnsNames, err := s.datavarehusAPI.GetTNSNames(ctx)
	if err != nil {
		return errs.E(op, fmt.Errorf("getting TNS names: %w", err))
	}

	foundTNSNames := map[string]struct{}{}
	for _, tnsName := range tnsNames {
		if slices.Contains(hosts, tnsName.Host) {
			foundTNSNames[strings.ToLower(tnsName.TnsName)] = struct{}{}
		}
	}

	if len(foundTNSNames) > 0 {
		if len(vm.IPs) != 1 {
			return errs.E(op, fmt.Errorf("for user %s expected exactly one IP, got %d", ident, len(vm.IPs)))
		}

		claims := &service.DVHClaims{
			Ident:               strings.ToLower(ident),
			IP:                  vm.IPs[0],
			Databases:           maps.Keys(foundTNSNames),
			Reference:           requestID,
			PodName:             s.podName,
			KnastContainerImage: config.Image,
		}

		signedJWT, err := s.iamcredentialsAPI.SignJWT(ctx, s.signerServiceAccount, claims.ToMapClaims())
		if err != nil {
			return errs.E(op, fmt.Errorf("signing JWT: %w", err))
		}

		err = s.datavarehusAPI.SendJWT(ctx, signedJWT.KeyID, signedJWT.SignedJWT)
		if err != nil {
			return errs.E(op, fmt.Errorf("sending JWT: %w", err))
		}
	}

	return nil
}

func (s *workstationService) AddWorkstationZonalTagBinding(ctx context.Context, zone, parent, tagNamespacedName string) error {
	const op errs.Op = "workstationService.AddWorkstationZonalTagBinding"

	err := s.cloudResourceManagerAPI.CreateZonalTagBinding(ctx, zone, parent, tagNamespacedName)
	if err != nil {
		return errs.E(op, fmt.Errorf("creating zonal tag binding: %w", err))
	}

	return nil
}

func (s *workstationService) RemoveWorkstationZonalTagBinding(ctx context.Context, zone, parent, tagValue string) error {
	const op errs.Op = "workstationService.RemoveWorkstationZonalTagBinding"

	err := s.cloudResourceManagerAPI.DeleteZonalTagBinding(ctx, zone, parent, tagValue)
	if err != nil {
		return errs.E(op, fmt.Errorf("deleting zonal tag binding: %w", err))
	}

	return nil
}

func (s *workstationService) StopWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.StopWorkstation"

	slug := user.Ident

	config, err := s.workstationAPI.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: slug,
	})
	if err != nil {
		return errs.E(op, err)
	}

	vms, err := s.computeAPI.GetVirtualMachinesByLabel(ctx, config.ReplicaZones, &service.Label{
		Key: service.WorkstationConfigIDLabel,
	})
	if err != nil {
		return errs.E(op, err)
	}

	err = s.workstationAPI.StopWorkstation(ctx, &service.WorkstationIdentifier{
		Slug:                  slug,
		WorkstationConfigSlug: slug,
	})
	if err != nil {
		return errs.E(op, err)
	}

	if len(vms) != 1 {
		return errs.E(op, fmt.Errorf("for user %s expected exactly one VM, got %d", slug, len(vms)))
	}

	err = s.workstationStorage.CreateWorkstationsActivity(ctx, slug, strconv.FormatUint(vms[0].ID, 10), service.WorkstationActionTypeStop)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *workstationService) EnsureWorkstation(ctx context.Context, user *service.User, input *service.WorkstationInput) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.EnsureWorkstation"

	slug := user.Ident

	sa, err := s.serviceAccountAPI.EnsureServiceAccount(ctx, &service.ServiceAccountRequest{
		ProjectID:   s.workstationsProject,
		AccountID:   service.WorkstationServiceAccountID(slug),
		DisplayName: slug,
		Description: fmt.Sprintf("Workstation service account for %s (%s)", user.Name, user.Email),
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring service account for %s: %w", user.Email, err))
	}

	err = s.serviceAccountAPI.AddServiceAccountPolicyBinding(ctx, s.workstationsProject, sa.Email, &service.Binding{
		Role: service.ServiceAccountUserRole,
		Members: []string{
			fmt.Sprintf("serviceAccount:%s", s.administratorServiceAcccount),
		},
	})
	if err != nil {
		return nil, err
	}

	err = s.artifactRegistryAPI.AddArtifactRegistryPolicyBinding(ctx, &service.ContainerRepositoryIdentifier{
		Project:    s.artifactRepositoryProject,
		Location:   s.location,
		Repository: s.artifactRepositoryName,
	}, &service.Binding{
		Role: service.ArtifactRegistryReaderRole,
		Members: []string{
			fmt.Sprintf("serviceAccount:%s", sa.Email),
		},
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring artifact registry policy binding: %w", err))
	}

	rules, err := s.computeAPI.GetFirewallRulesForRegionalPolicy(ctx, s.workstationsProject, s.location, s.firewallPolicyName)
	if err != nil {
		return nil, errs.E(op, err)
	}

	allowedHosts := map[string]struct{}{}
	for _, rule := range rules {
		allowedHosts[rule.Name] = struct{}{}
	}

	c, w, err := s.workstationAPI.EnsureWorkstationWithConfig(ctx, &service.EnsureWorkstationOpts{
		Workstation: service.WorkstationOpts{
			Slug:        slug,
			ConfigName:  slug,
			DisplayName: displayName(user),
			Labels:      service.DefaultWorkstationLabels(slug),
		},
		Config: service.WorkstationConfigOpts{
			Slug:                slug,
			DisplayName:         displayName(user),
			MachineType:         input.MachineType,
			ServiceAccountEmail: sa.Email,
			SubjectEmail:        user.Email,
			Labels:              service.DefaultWorkstationLabels(slug),
			Env:                 service.DefaultWorkstationEnv(slug, user.Email, user.Name),
			ContainerImage:      input.ContainerImage,
		},
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring workstation for %s: %w", user.Email, err))
	}

	err = s.workstationAPI.AddWorkstationUser(
		ctx,
		&service.WorkstationIdentifier{
			Slug:                  slug,
			WorkstationConfigSlug: slug,
		},
		user.Email,
	)
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("adding user to workstation %s: %w", user.Email, err))
	}

	err = s.secureWebProxyAPI.EnsureSecurityPolicyRuleWithRandomPriority(ctx, &service.PolicyRuleEnsureNextAvailablePortOpts{
		ID: &service.PolicyIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Policy:   s.tlsSecureWebProxyPolicy,
		},
		PriorityMinRange:     service.FirewallDenyRulePriorityMin,
		PriorityMaxRange:     service.FirewallDenyRulePriorityMax,
		BasicProfile:         "DENY",
		Description:          fmt.Sprintf("Default deny secure policy rule for workstation user %s ", displayName(user)),
		Enabled:              true,
		Name:                 slug,
		SessionMatcher:       createSessionMatch(sa.Email),
		TlsInspectionEnabled: true,
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring workstation default deny secure policy rule for %s: %w", user.Email, err))
	}

	urlList, err := s.workstationStorage.GetLastWorkstationsURLList(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("getting workstation url list for %s: %w", user.Email, err))
	}

	err = s.secureWebProxyAPI.EnsureURLList(ctx, &service.URLListEnsureOpts{
		ID: &service.URLListIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Slug:     slug,
		},
		Description: fmt.Sprintf("URL list for user %s ", displayName(user)),
		URLS:        urlList.URLAllowList,
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring workstation urllist for %s: %w", user.Email, err))
	}

	err = s.secureWebProxyAPI.EnsureSecurityPolicyRuleWithRandomPriority(ctx, &service.PolicyRuleEnsureNextAvailablePortOpts{
		ID: &service.PolicyIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Policy:   s.tlsSecureWebProxyPolicy,
		},
		PriorityMinRange:     service.FirewallAllowRulePriorityMin,
		PriorityMaxRange:     service.FirewallAllowRulePriorityMax,
		ApplicationMatcher:   createApplicationMatch(s.workstationsProject, s.location, slug),
		BasicProfile:         "ALLOW",
		Description:          fmt.Sprintf("Secure policy rule for workstation user %s ", displayName(user)),
		Enabled:              true,
		Name:                 normalize.Email("allow-" + slug),
		SessionMatcher:       createSessionMatch(sa.Email),
		TlsInspectionEnabled: true,
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring workstation secure policy rule for %s: %w", user.Email, err))
	}

	if !urlList.DisableGlobalAllowList {
		err = s.secureWebProxyAPI.EnsureSecurityPolicyRuleWithRandomPriority(ctx, &service.PolicyRuleEnsureNextAvailablePortOpts{
			ID: &service.PolicyIdentifier{
				Project:  s.workstationsProject,
				Location: s.location,
				Policy:   s.tlsSecureWebProxyPolicy,
			},
			PriorityMinRange:     service.FirewallAllowRulePriorityMin,
			PriorityMaxRange:     service.FirewallAllowRulePriorityMax,
			ApplicationMatcher:   createApplicationMatch(s.workstationsProject, s.location, service.GlobalURLAllowListName),
			BasicProfile:         "ALLOW",
			Description:          fmt.Sprintf("Secure policy rule for workstation user %s ", displayName(user)),
			Enabled:              true,
			Name:                 normalize.Email("global-allow-" + slug),
			SessionMatcher:       createSessionMatch(sa.Email),
			TlsInspectionEnabled: true,
		})
		if err != nil {
			return nil, errs.E(op, fmt.Errorf("ensuring workstation secure policy rule for %s: %w", user.Email, err))
		}
	}

	if urlList.DisableGlobalAllowList {
		err = s.secureWebProxyAPI.DeleteSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Policy:   s.tlsSecureWebProxyPolicy,
			Slug:     normalize.Email("global-allow-" + slug),
		})
		if err != nil {
			return nil, errs.E(op, fmt.Errorf("delete security policy rule for workstation user %s: %w", user.Email, err))
		}
	}

	return &service.WorkstationOutput{
		Slug:        w.Slug,
		DisplayName: w.DisplayName,
		Reconciling: w.Reconciling,
		CreateTime:  w.CreateTime,
		UpdateTime:  w.UpdateTime,
		StartTime:   w.StartTime,
		State:       w.State,
		Config: &service.WorkstationConfigOutput{
			CreateTime:     c.CreateTime,
			UpdateTime:     c.UpdateTime,
			IdleTimeout:    c.IdleTimeout,
			RunningTimeout: c.RunningTimeout,
			MachineType:    c.MachineType,
			Image:          c.Image,
			Env:            c.Env,
		},
		Host: w.Host,
	}, nil
}

func (s *workstationService) GetWorkstation(ctx context.Context, user *service.User) (*service.WorkstationOutput, error) {
	return s.GetWorkstationBySlug(ctx, user.Ident)
}

func (s *workstationService) GetWorkstationBySlug(ctx context.Context, slug string) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.GetWorkstation"

	c, err := s.workstationAPI.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: slug,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	w, err := s.workstationAPI.GetWorkstation(ctx, &service.WorkstationIdentifier{
		Slug:                  slug,
		WorkstationConfigSlug: slug,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationOutput{
		Slug:        w.Slug,
		DisplayName: w.DisplayName,
		Reconciling: w.Reconciling,
		CreateTime:  w.CreateTime,
		UpdateTime:  w.UpdateTime,
		StartTime:   w.StartTime,
		State:       w.State,
		Config: &service.WorkstationConfigOutput{
			CreateTime:     c.CreateTime,
			UpdateTime:     c.UpdateTime,
			IdleTimeout:    c.IdleTimeout,
			RunningTimeout: c.RunningTimeout,
			MachineType:    c.MachineType,
			Image:          c.Image,
			Env:            c.Env,
		},
		Host: w.Host,
	}, nil
}

func (s *workstationService) DeleteWorkstationByUser(ctx context.Context, user *service.User) error {
	slug := user.Ident
	return s.DeleteWorkstationBySlug(ctx, slug)
}

func (s *workstationService) DeleteUrlAllowList(ctx context.Context, urlList *service.URLListIdentifier) error {
	for retry := 0; retry < 10; retry++ {
		var gapierr *googleapi.Error
		err := s.secureWebProxyAPI.DeleteURLList(ctx, urlList)
		if err == nil {
			break
		} else if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			break
		} else if gapierr.Code == http.StatusBadRequest {
			time.Sleep(1 * time.Second)
		} else {
			return err
		}
	}
	return nil
}

func (s *workstationService) DeleteWorkstationBySlug(ctx context.Context, slug string) error {
	const op errs.Op = "workstationService.DeleteWorkstationBySlug"

	err := s.secureWebProxyAPI.DeleteSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Policy:   s.tlsSecureWebProxyPolicy,
		Slug:     slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete security policy deny rule for workstation user: %w", err))
	}

	err = s.secureWebProxyAPI.DeleteSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Policy:   s.tlsSecureWebProxyPolicy,
		Slug:     normalize.Email("allow-" + slug),
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete security policy allow rule for workstation user: %w", err))
	}

	err = s.secureWebProxyAPI.DeleteSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Policy:   s.tlsSecureWebProxyPolicy,
		Slug:     normalize.Email("global-allow-" + slug),
	})

	var gapierr *googleapi.Error
	if err != nil && !(errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound) {
		return errs.E(op, fmt.Errorf("delete security policy allow rule for workstation user: %w", err))
	}

	err = s.DeleteUrlAllowList(ctx, &service.URLListIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Slug:     slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete urllist for workstation user: %w", err))
	}

	// FIXME: create and delete should expect the same input
	err = s.serviceAccountAPI.DeleteServiceAccount(ctx, s.serviceAccountsProject, serviceAccountEmail(s.serviceAccountsProject, slug))
	if err != nil {
		return errs.E(op, fmt.Errorf("delete workstation service account for user: %w", err))
	}

	err = s.workstationAPI.DeleteWorkstationConfig(context.Background(), &service.WorkstationConfigDeleteOpts{
		Slug:   slug,
		NoWait: true,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete workstation config for user: %w", err))
	}

	return nil
}

func (s *workstationService) ListWorkstations(ctx context.Context) ([]*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.ListWorkstations"

	wcs, err := s.workstationAPI.ListWorkstationConfigs(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	cws := make([]*service.WorkstationOutput, 0, len(wcs))
	for _, wc := range wcs {
		w, err := s.GetWorkstationBySlug(ctx, wc.Slug)
		if err == nil {
			cws = append(cws, w)
		}
	}
	return cws, nil
}

func (s *workstationService) UpdateWorkstationURLList(ctx context.Context, user *service.User, input *service.WorkstationURLList) error {
	const op errs.Op = "workstationService.UpdateWorkstationURLList"

	slug := user.Ident

	urlList := input.URLAllowList
	uniqueURLList := make(map[string]struct{})
	for _, u := range urlList {
		if len(u) == 0 {
			continue
		}

		uniqueURLList[u] = struct{}{}
	}

	urlList = maps.Keys(uniqueURLList)
	slices.Sort(urlList)

	err := s.secureWebProxyAPI.UpdateURLList(ctx, &service.URLListUpdateOpts{
		ID: &service.URLListIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Slug:     slug,
		},
		URLS: urlList,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("updating workstation urllist for %s: %w", user.Email, err))
	}

	saEmail := service.ServiceAccountEmailFromAccountID(s.workstationsProject, service.WorkstationServiceAccountID(slug))

	if !input.DisableGlobalAllowList {
		err = s.secureWebProxyAPI.EnsureSecurityPolicyRuleWithRandomPriority(ctx, &service.PolicyRuleEnsureNextAvailablePortOpts{
			ID: &service.PolicyIdentifier{
				Project:  s.workstationsProject,
				Location: s.location,
				Policy:   s.tlsSecureWebProxyPolicy,
			},
			PriorityMinRange:     service.FirewallAllowRulePriorityMin,
			PriorityMaxRange:     service.FirewallAllowRulePriorityMax,
			ApplicationMatcher:   createApplicationMatch(s.workstationsProject, s.location, service.GlobalURLAllowListName),
			BasicProfile:         "ALLOW",
			Description:          fmt.Sprintf("Secure policy rule for workstation user %s ", displayName(user)),
			Enabled:              true,
			Name:                 normalize.Email("global-allow-" + slug),
			SessionMatcher:       createSessionMatch(saEmail),
			TlsInspectionEnabled: true,
		})
		if err != nil {
			return errs.E(op, fmt.Errorf("ensuring workstation secure policy rule for %s: %w", user.Email, err))
		}
	}

	if input.DisableGlobalAllowList {
		err = s.secureWebProxyAPI.DeleteSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Policy:   s.tlsSecureWebProxyPolicy,
			Slug:     normalize.Email("global-allow-" + slug),
		})
		if err != nil {
			return errs.E(op, fmt.Errorf("delete security policy rule for workstation user %s: %w", user.Email, err))
		}
	}

	err = s.workstationStorage.CreateWorkstationsURLListChange(ctx, user.Ident, input)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func displayName(user *service.User) string {
	return fmt.Sprintf("%s (%s)", user.Name, user.Email)
}

func serviceAccountEmail(project, userEmail string) string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", normalize.Email(userEmail), project)
}

func createSessionMatch(saEmail string) string {
	return fmt.Sprintf("source.matchServiceAccount('%s')", saEmail)
}

func createApplicationMatch(project, location, slug string) string {
	return fmt.Sprintf("inUrlList(request.url(), 'projects/%v/locations/%s/urlLists/%s')", project, location, slug)
}

func NewWorkstationService(
	workstationsProject string,
	serviceAccountsProject string,
	location string,
	tlsSecureWebProxyPolicy string,
	firewallPolicyName string,
	loggingBucket string,
	loggingView string,
	administratorServiceAcccount string,
	artifactRepositoryName string,
	artifactRepositoryProject string,
	signerServiceAccount string,
	podName string,
	serviceAccountAPI service.ServiceAccountAPI,
	crm service.CloudResourceManagerAPI,
	swp service.SecureWebProxyAPI,
	workstationsAPI service.WorkstationsAPI,
	workstationsStorage service.WorkstationsStorage,
	computeAPI service.ComputeAPI,
	clapi service.CloudLoggingAPI,
	store service.WorkstationsQueue,
	artifactRegistryAPI service.ArtifactRegistryAPI,
	datavarehusAPI service.DatavarehusAPI,
	iamcredentialsAPI service.IAMCredentialsAPI,
	cloudBillingAPI service.CloudBillingAPI,
) *workstationService {
	return &workstationService{
		workstationsProject:          workstationsProject,
		workstationsLoggingBucket:    loggingBucket,
		workstationsLoggingView:      loggingView,
		serviceAccountsProject:       serviceAccountsProject,
		location:                     location,
		tlsSecureWebProxyPolicy:      tlsSecureWebProxyPolicy,
		firewallPolicyName:           firewallPolicyName,
		administratorServiceAcccount: administratorServiceAcccount,
		artifactRepositoryName:       artifactRepositoryName,
		artifactRepositoryProject:    artifactRepositoryProject,
		signerServiceAccount:         signerServiceAccount,
		podName:                      podName,
		workstationsQueue:            store,
		workstationStorage:           workstationsStorage,
		workstationAPI:               workstationsAPI,
		serviceAccountAPI:            serviceAccountAPI,
		secureWebProxyAPI:            swp,
		cloudResourceManagerAPI:      crm,
		computeAPI:                   computeAPI,
		cloudLoggingAPI:              clapi,
		artifactRegistryAPI:          artifactRegistryAPI,
		datavarehusAPI:               datavarehusAPI,
		iamcredentialsAPI:            iamcredentialsAPI,
		cloudBillingAPI:              cloudBillingAPI,
	}
}
