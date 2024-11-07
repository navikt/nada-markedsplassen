package core

import (
	"context"
	"fmt"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/navikt/nada-backend/pkg/normalize"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
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

	workstationsStorage     service.WorkstationsStorage
	workstationAPI          service.WorkstationsAPI
	serviceAccountAPI       service.ServiceAccountAPI
	secureWebProxyAPI       service.SecureWebProxyAPI
	cloudResourceManagerAPI service.CloudResourceManagerAPI
	computeAPI              service.ComputeAPI
	cloudLoggingAPI         service.CloudLoggingAPI
	artifactRegistryAPI     service.ArtifactRegistryAPI
}

func (s *workstationService) CreateWorkstationStartJob(ctx context.Context, user *service.User) (*service.WorkstationStartJob, error) {
	const op errs.Op = "workstationService.CreateWorkstationStartJob"

	_, err := s.GetWorkstation(ctx, user)
	if err != nil {
		return nil, err
	}

	job, err := s.workstationsStorage.CreateWorkstationStartJob(ctx, user.Ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return job, nil
}

func (s *workstationService) GetWorkstationStartJobsForUser(ctx context.Context, ident string) (*service.WorkstationStartJobs, error) {
	const op errs.Op = "workstationService.GetWorkstationStartJobsForUser"

	jobs, err := s.workstationsStorage.GetWorkstationStartJobsForUser(ctx, ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationStartJobs{
		Jobs: jobs,
	}, nil
}

func (s *workstationService) GetWorkstationJobsForUser(ctx context.Context, ident string) (*service.WorkstationJobs, error) {
	const op errs.Op = "workstationService.GetRunningWorkstationJobsForUser"

	jobs, err := s.workstationsStorage.GetWorkstationJobsForUser(ctx, ident)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.WorkstationJobs{
		Jobs: jobs,
	}, nil
}

func (s *workstationService) CreateWorkstationJob(ctx context.Context, user *service.User, input *service.WorkstationInput) (*service.WorkstationJob, error) {
	const op errs.Op = "workstationService.CreateWorkstationJob"

	job, err := s.workstationsStorage.CreateWorkstationJob(ctx, &service.WorkstationJobOpts{
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

	job, err := s.workstationsStorage.GetWorkstationJob(ctx, jobID)
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

	raw, err := s.computeAPI.GetFirewallRulesForRegionalPolicy(ctx, s.workstationsProject, s.location, s.firewallPolicyName)
	if err != nil {
		return nil, errs.E(op, err)
	}

	var tags []*service.FirewallTag
	for _, rule := range raw {
		// Filter out the default deny and allow rules, which do not have securetags
		if rule.SecureTags == nil || len(rule.SecureTags) != 1 {
			continue
		}

		tags = append(tags, &service.FirewallTag{
			Name: rule.Name,
			// This is fragile, but we know that there aren't more than one securetag per rule
			SecureTag: rule.SecureTags[0],
		})
	}

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
			Image:       image.URI,
			Description: image.Name,
			Labels:      image.Manifest.Labels,
		}
	}

	return &service.WorkstationOptions{
		FirewallTags:        tags,
		MachineTypes:        service.WorkstationMachineTypes(),
		ContainerImages:     containerImages,
		DefaultURLAllowList: service.DefaultURLAllowList(),
	}, nil
}

const workstationConfigID = "workstation_config_id"

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
		return errs.E(op, err)
	}

	vms, err := s.computeAPI.GetVirtualMachinesByLabel(ctx, config.ReplicaZones, &service.Label{
		Key:   workstationConfigID,
		Value: slug,
	})
	if err != nil {
		return errs.E(op, err)
	}

	for _, vm := range vms {
		value, hasAllowlist := config.Annotations[service.WorkstationOnpremAllowlistAnnotation]
		if !hasAllowlist || value == "" {
			continue
		}

		for _, host := range strings.Split(value, ",") {
			err := s.cloudResourceManagerAPI.CreateZonalTagBinding(
				ctx,
				vm.Zone,
				fmt.Sprintf("//compute.googleapis.com/projects/%s/zones/%s/instances/%d", s.workstationsProject, vm.Zone, vm.ID),
				fmt.Sprintf("%s/%s/%s", s.workstationsProject, host, host),
			)
			if err != nil {
				return errs.E(op, err)
			}
		}
	}

	return nil
}

func (s *workstationService) StopWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.StopWorkstation"

	slug := user.Ident

	err := s.workstationAPI.StopWorkstation(ctx, &service.WorkstationIdentifier{
		Slug:                  slug,
		WorkstationConfigSlug: slug,
	})
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

	for _, host := range input.OnPremAllowList {
		if _, ok := allowedHosts[host]; !ok {
			return nil, errs.E(errs.Invalid, service.CodeUnknownHostInOnPremAllowList, op, fmt.Errorf("on-prem allow list contains unknown host: %s", host))
		}
	}

	var onPremHosts []string
	var annotations = make(map[string]string)
	for _, v := range input.OnPremAllowList {
		if v == "" {
			continue
		}
		onPremHosts = append(onPremHosts, v)
	}
	if len(onPremHosts) > 0 {
		annotations[service.WorkstationOnpremAllowlistAnnotation] = strings.Join(onPremHosts, ",")
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
			Annotations: map[string]string{
				service.WorkstationOnpremAllowlistAnnotation: strings.Join(input.OnPremAllowList, ","),
			},
			Labels:         service.DefaultWorkstationLabels(slug),
			Env:            service.DefaultWorkstationEnv(slug, user.Email, user.Name),
			ContainerImage: input.ContainerImage,
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

	urlList := append(input.URLAllowList, service.DefaultURLAllowList()...)
	uniqueURLList := make(map[string]struct{})
	for _, u := range urlList {
		uniqueURLList[u] = struct{}{}
	}

	urlList = maps.Keys(uniqueURLList)
	slices.Sort(urlList)

	err = s.secureWebProxyAPI.EnsureURLList(ctx, &service.URLListEnsureOpts{
		ID: &service.URLListIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Slug:     slug,
		},
		Description: fmt.Sprintf("URL list for user %s ", displayName(user)),
		URLS:        urlList,
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
		URLAllowList: input.URLAllowList,
		Host:         w.Host,
	}, nil
}

func (s *workstationService) GetWorkstation(ctx context.Context, user *service.User) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.GetWorkstation"

	slug := user.Ident

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

	urlList, err := s.secureWebProxyAPI.GetURLList(ctx, &service.URLListIdentifier{Slug: slug, Project: s.workstationsProject, Location: s.location})
	if err != nil {
		return nil, errs.E(op, err)
	}

	firewallRulesAllowList := []string{}
	if rules, ok := c.Annotations[service.WorkstationOnpremAllowlistAnnotation]; ok && rules != "" {
		firewallRulesAllowList = strings.Split(rules, ",")
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
			CreateTime:             c.CreateTime,
			UpdateTime:             c.UpdateTime,
			IdleTimeout:            c.IdleTimeout,
			RunningTimeout:         c.RunningTimeout,
			FirewallRulesAllowList: firewallRulesAllowList,
			MachineType:            c.MachineType,
			Image:                  c.Image,
			Env:                    c.Env,
		},
		URLAllowList: urlList,
		Host:         w.Host,
	}, nil
}

func (s *workstationService) DeleteWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.DeleteWorkstation"

	slug := user.Ident

	err := s.secureWebProxyAPI.DeleteSecurityPolicyRule(ctx, &service.PolicyRuleIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Policy:   s.tlsSecureWebProxyPolicy,
		Slug:     slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete security policy rule for workstation user %s: %w", user.Email, err))
	}

	err = s.secureWebProxyAPI.DeleteURLList(ctx, &service.URLListIdentifier{
		Project:  s.workstationsProject,
		Location: s.location,
		Slug:     slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete urllist for workstation user %s: %w", user.Email, err))
	}

	err = s.workstationAPI.DeleteWorkstationConfig(ctx, &service.WorkstationConfigDeleteOpts{
		Slug: slug,
	})
	if err != nil {
		return errs.E(op, fmt.Errorf("delete workstation config for user %s: %w", user.Email, err))
	}

	// FIXME: create and delete should expect the same input
	err = s.serviceAccountAPI.DeleteServiceAccount(ctx, s.serviceAccountsProject, serviceAccountEmail(s.serviceAccountsProject, user.Email))
	if err != nil {
		return errs.E(op, fmt.Errorf("delete workstation service account for user %s: %w", user.Email, err))
	}

	return nil
}

func (s *workstationService) UpdateWorkstationURLList(ctx context.Context, user *service.User, input *service.WorkstationURLList) error {
	const op errs.Op = "workstationService.UpdateWorkstationURLList"

	slug := user.Ident

	urlList := append(input.URLAllowList, service.DefaultURLAllowList()...)
	uniqueURLList := make(map[string]struct{})
	for _, u := range urlList {
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
	s service.ServiceAccountAPI,
	crm service.CloudResourceManagerAPI,
	swp service.SecureWebProxyAPI,
	w service.WorkstationsAPI,
	computeAPI service.ComputeAPI,
	clapi service.CloudLoggingAPI,
	store service.WorkstationsStorage,
	artifactRegistryAPI service.ArtifactRegistryAPI,
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
		workstationAPI:               w,
		serviceAccountAPI:            s,
		secureWebProxyAPI:            swp,
		cloudResourceManagerAPI:      crm,
		computeAPI:                   computeAPI,
		cloudLoggingAPI:              clapi,
		workstationsStorage:          store,
		artifactRegistryAPI:          artifactRegistryAPI,
	}
}
