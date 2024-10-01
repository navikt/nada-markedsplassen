package core

import (
	"context"
	"fmt"

	"github.com/navikt/nada-backend/pkg/normalize"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.WorkstationsService = (*workstationService)(nil)

type workstationService struct {
	workstationsProject     string
	serviceAccountsProject  string
	location                string
	tlsSecureWebProxyPolicy string

	workstationAPI          service.WorkstationsAPI
	serviceAccountAPI       service.ServiceAccountAPI
	secureWebProxyAPI       service.SecureWebProxyAPI
	cloudResourceManagerAPI service.CloudResourceManagerAPI
}

func (s *workstationService) StartWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.StartWorkstation"

	slug := normalize.Email(user.Email)

	err := s.workstationAPI.StartWorkstation(ctx, &service.WorkstationStartOpts{
		Slug:       slug,
		ConfigName: slug,
	})
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *workstationService) StopWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.StopWorkstation"

	slug := normalize.Email(user.Email)

	err := s.workstationAPI.StopWorkstation(ctx, &service.WorkstationStopOpts{
		Slug:       slug,
		ConfigName: slug,
	})
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *workstationService) EnsureWorkstation(ctx context.Context, user *service.User, input *service.WorkstationInput) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.EnsureWorkstation"

	slug := normalize.Email(user.Email)

	sa, err := s.serviceAccountAPI.EnsureServiceAccount(ctx, &service.ServiceAccountRequest{
		ProjectID:   s.serviceAccountsProject,
		AccountID:   slug,
		DisplayName: slug,
		Description: fmt.Sprintf("Workstation service account for %s (%s)", user.Name, user.Email),
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring service account for %s: %w", user.Email, err))
	}

	err = s.cloudResourceManagerAPI.AddProjectIAMPolicyBinding(ctx, s.workstationsProject, &service.Binding{
		Role: service.WorkstationOperationViewerRole(s.workstationsProject),
		Members: []string{
			fmt.Sprintf("user:%s", user.Email),
		},
	})
	if err != nil {
		return nil, errs.E(op, err)
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
			ServiceAccountEmail: sa.Email,
			SubjectEmail:        user.Email,
			Labels:              service.DefaultWorkstationLabels(slug),
			MachineType:         input.MachineType,
			ContainerImage:      input.ContainerImage,
		},
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring workstation for %s: %w", user.Email, err))
	}

	err = s.secureWebProxyAPI.EnsureURLList(ctx, &service.URLListEnsureOpts{
		ID: &service.URLListIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Slug:     slug,
		},
		Description: fmt.Sprintf("URL list for user %s ", displayName(user)),
		URLS:        input.URLAllowList,
	})
	if err != nil {
		return nil, errs.E(op, fmt.Errorf("ensuring workstation urllist for %s: %w", user.Email, err))
	}

	err = s.secureWebProxyAPI.EnsureSecurityPolicyRule(ctx, &service.PolicyRuleEnsureOpts{
		ID: &service.PolicyRuleIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Policy:   s.tlsSecureWebProxyPolicy,
			Slug:     slug,
		},
		Rule: &service.GatewaySecurityPolicyRule{
			SessionMatcher:       createSessionMatch(sa.Email),
			ApplicationMatcher:   createApplicationMatch(s.workstationsProject, s.location, slug),
			BasicProfile:         "ALLOW",
			Description:          fmt.Sprintf("Secure policy rule for workstation user %s ", displayName(user)),
			Enabled:              true,
			Name:                 slug,
			Priority:             100,
			TlsInspectionEnabled: true,
		},
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
	}, nil
}

func (s *workstationService) GetWorkstation(ctx context.Context, user *service.User) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.GetWorkstation"

	slug := normalize.Email(user.Email)

	c, err := s.workstationAPI.GetWorkstationConfig(ctx, &service.WorkstationConfigGetOpts{
		Slug: slug,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	w, err := s.workstationAPI.GetWorkstation(ctx, &service.WorkstationGetOpts{
		Slug:       slug,
		ConfigName: slug,
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	urlList, err := s.secureWebProxyAPI.GetURLList(ctx, &service.URLListIdentifier{Slug: slug, Project: s.workstationsProject, Location: s.location})
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
		URLAllowList: urlList,
	}, nil
}

func (s *workstationService) DeleteWorkstation(ctx context.Context, user *service.User) error {
	const op errs.Op = "workstationService.DeleteWorkstation"

	slug := normalize.Email(user.Email)

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

	err = s.cloudResourceManagerAPI.RemoveProjectIAMPolicyBindingMemberForRole(
		ctx,
		s.workstationsProject,
		service.WorkstationOperationViewerRole(s.workstationsProject),
		fmt.Sprintf("user:%s", user.Email),
	)
	if err != nil {
		return errs.E(op, err)
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

	slug := normalize.Email(user.Email)

	err := s.secureWebProxyAPI.UpdateURLList(ctx, &service.URLListUpdateOpts{
		ID: &service.URLListIdentifier{
			Project:  s.workstationsProject,
			Location: s.location,
			Slug:     slug,
		},
		URLS: input.URLAllowList,
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

func NewWorkstationService(workstationsProject, serviceAccountsProject, location, tlsSecureWebProxyPolicy string, s service.ServiceAccountAPI, crm service.CloudResourceManagerAPI, swp service.SecureWebProxyAPI, w service.WorkstationsAPI) *workstationService {
	return &workstationService{
		workstationsProject:     workstationsProject,
		serviceAccountsProject:  serviceAccountsProject,
		location:                location,
		tlsSecureWebProxyPolicy: tlsSecureWebProxyPolicy,
		serviceAccountAPI:       s,
		secureWebProxyAPI:       swp,
		cloudResourceManagerAPI: crm,
		workstationAPI:          w,
	}
}
