package core

import (
	"context"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type workstationService struct {
	projectID string

	workstationAPI    service.WorkstationsAPI
	serviceAccountAPI service.ServiceAccountAPI
}

func (s *workstationService) EnsureWorkstation(ctx context.Context, user *service.User, input *service.WorkstationInput) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.EnsureWorkstation"

	slug := NormalizedEmail(user.Email)

	sa, err := s.serviceAccountAPI.EnsureServiceAccount(ctx, &service.ServiceAccountRequest{
		ProjectID:   s.projectID,
		AccountID:   slug,
		DisplayName: slug,
		Description: fmt.Sprintf("Workstation service account for %s (%s)", user.Name, user.Email),
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	// FIXME: Need to grant the correct roles for the user to able to access the created workstation

	err = s.workstationAPI.EnsureWorkstationWithConfig(ctx, &service.EnsureWorkstationOpts{
		Workstation: service.WorkstationOpts{
			Slug:        slug,
			ConfigName:  slug,
			DisplayName: displayName(user),
			Labels:      service.DefaultWorkstationLabels(user.Email),
		},
		Config: service.WorkstationConfigOpts{
			Slug:                slug,
			DisplayName:         displayName(user),
			ServiceAccountEmail: sa.Email,
			SubjectEmail:        user.Email,
			Labels:              service.DefaultWorkstationLabels(user.Email),
			MachineType:         input.MachineType,
			ContainerImage:      input.ContainerImage,
		},
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *workstationService) GetWorkstation(ctx context.Context, user *service.User) (*service.WorkstationOutput, error) {
	const op errs.Op = "workstationService.GetWorkstation"

	slug := NormalizedEmail(user.Email)

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

	return &service.WorkstationOutput{
		Workstation:       w,
		WorkstationConfig: c,
	}, nil
}

func (s *workstationService) DeleteWorkstation(ctx context.Context, user *service.User) error {
	return nil
}

func NormalizedEmail(email string) string {
	slug.MaxLength = 63
	slug.CustomSub = map[string]string{
		"_": "-",
		"@": "-at-",
	}

	return slug.Make(email)
}

func displayName(user *service.User) string {
	return fmt.Sprintf("%s (%s)", user.Name, user.Email)
}

func NewWorkstationService(projectID string, s service.ServiceAccountAPI, w service.WorkstationsAPI) *workstationService {
	return &workstationService{
		projectID:         projectID,
		workstationAPI:    w,
		serviceAccountAPI: s,
	}
}
