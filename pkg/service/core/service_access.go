package core

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.AccessService = (*accessService)(nil)

type accessService struct {
	dataCatalogueURL    string
	allUsersEmail       string
	slackapi            service.SlackAPI
	pollyStorage        service.PollyStorage
	accessStorage       service.AccessStorage
	dataProductStorage  service.DataProductsStorage
	bigQueryStorage     service.BigQueryStorage
	joinableViewStorage service.JoinableViewsStorage
	bigQueryAPI         service.BigQueryAPI
	dataProductService  service.DataProductsService
	metabaseService     service.MetabaseService
}

func (s *accessService) GetAccessRequests(ctx context.Context, datasetID uuid.UUID) (*service.AccessRequestsWrapper, error) {
	const op errs.Op = "accessService.GetAccessRequests"

	requests, err := s.accessStorage.ListAccessRequestsForDataset(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	for i := range requests {
		r := &requests[i]
		polly := &service.Polly{}
		if r.Polly != nil {
			polly, err = s.pollyStorage.GetPollyDocumentation(ctx, r.Polly.ID)
			if err != nil {
				return nil, errs.E(op, err)
			}
		}
		r.Polly = polly
	}

	return &service.AccessRequestsWrapper{
		AccessRequests: requests,
	}, nil
}

// nolint: cyclop
func (s *accessService) CreateAccessRequest(ctx context.Context, user *service.User, input service.NewAccessRequestDTO) error {
	const op errs.Op = "accessService.CreateAccessRequest"

	identity := newResolvedAccessIdentity(user, accessTarget{
		input.Subject,
		input.SubjectType,
		input.Owner,
	})

	var pollyID uuid.NullUUID

	if input.Polly != nil {
		dbPolly, err := s.pollyStorage.CreatePollyDocumentation(ctx, *input.Polly)
		if err != nil {
			return errs.E(op, err)
		}

		pollyID = uuid.NullUUID{UUID: dbPolly.ID, Valid: true}
	}

	accessRequest, err := s.accessStorage.CreateAccessRequestForDataset(ctx, input.DatasetID, pollyID, identity.SubjectWithType(), identity.OwnerWithType(), input.Platform, input.Expires)
	if err != nil {
		return errs.E(op, err)
	}

	ds, err := s.dataProductStorage.GetDataset(ctx, input.DatasetID)
	if err != nil {
		return errs.E(op, err)
	}

	dp, err := s.dataProductStorage.GetDataproduct(ctx, ds.DataproductID)
	if err != nil {
		return errs.E(op, err)
	}

	slackMessage := createAccessRequestSlackNotification(dp, ds, s.dataCatalogueURL, accessRequest.Owner)

	if dp.Owner.TeamContact == nil || *dp.Owner.TeamContact == "" {
		return nil
	}

	err = s.slackapi.SendSlackNotification(*dp.Owner.TeamContact, slackMessage)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func createAccessRequestSlackNotification(dp *service.DataproductWithDataset, ds *service.Dataset, dataCatalogueURL, subject string) string {
	link := fmt.Sprintf(
		"\nLink: %s/dataproduct/%s/%s/%s",
		dataCatalogueURL,
		dp.ID.String(),
		url.QueryEscape(dp.Name),
		ds.ID.String(),
	)

	dsp := fmt.Sprintf(
		"\nDatasett: %s\nDataprodukt: %s",
		ds.Name,
		dp.Name,
	)

	return fmt.Sprintf(
		"%s har sendt en søknad om tilgang for: %s%s",
		subject,
		dsp,
		link,
	)
}

func (s *accessService) DeleteAccessRequest(ctx context.Context, user *service.User, accessRequestID uuid.UUID) error {
	const op errs.Op = "accessService.DeleteAccessRequest"

	accessRequest, err := s.accessStorage.GetAccessRequest(ctx, accessRequestID)
	if err != nil {
		return errs.E(op, err)
	}

	if err := ensureOwner(user, accessRequest.Owner); err != nil {
		return errs.E(op, err)
	}

	if err := s.accessStorage.DeleteAccessRequest(ctx, accessRequestID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) UpdateAccessRequest(ctx context.Context, input service.UpdateAccessRequestDTO) error {
	const op errs.Op = "accessService.UpdateAccessRequest"

	// FIXME: Should we allow updating without checking the owner?

	if input.Polly != nil {
		if input.Polly.ID == nil {
			dbPolly, err := s.pollyStorage.CreatePollyDocumentation(ctx, *input.Polly)
			if err != nil {
				return errs.E(op, err)
			}

			input.Polly.ID = &dbPolly.ID
		}
	}

	err := s.accessStorage.UpdateAccessRequest(ctx, input)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) ApproveAccessRequest(ctx context.Context, user *service.User, accessRequestID uuid.UUID) error {
	const op errs.Op = "accessService.ApproveAccessRequest"

	ar, err := s.accessStorage.GetAccessRequest(ctx, accessRequestID)
	if err != nil {
		return errs.E(op, err)
	}

	ds, err := s.dataProductStorage.GetDataset(ctx, ar.DatasetID)
	if err != nil {
		return errs.E(op, err)
	}

	if err = s.validateAccessGrant(ds, ar.Subject); err != nil {
		return errs.E(op, err)
	}

	identity := newResolvedAccessIdentity(user, accessTarget{
		&ar.Subject,
		&ar.SubjectType,
		&ar.Owner,
	})

	// TODO: Split ApproveAccessRequest in 2?
	switch ar.Platform {
	case service.AccessPlatformBigQuery:
		bq, err := s.bigQueryStorage.GetBigqueryDatasource(ctx, ds.ID, false)
		if err != nil {
			return errs.E(op, err)
		}
		err = s.grantBigQueryAccess(ctx, identity, ds.ID, bq.ProjectID)
		if err != nil {
			return errs.E(op, err)
		}
	case service.AccessPlatformMetabase:
		err = s.metabaseService.GrantMetabaseAccessRestricted(ctx, ds.ID, identity.Subject, identity.SubjectType)
		if err != nil {
			return errs.E(op, err)
		}
	}

	err = s.accessStorage.GrantAccessToDatasetAndApproveRequest(
		ctx,
		user,
		identity.SubjectWithType(),
		ar,
	)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) DenyAccessRequest(ctx context.Context, user *service.User, accessRequestID uuid.UUID, reason *string) error {
	const op errs.Op = "accessService.DenyAccessRequest"

	err := s.accessStorage.DenyAccessRequest(ctx, user, accessRequestID, reason)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

// nolint: cyclop
func (s *accessService) RevokeBigQueryAccessToDataset(ctx context.Context, access *service.Access, gcpProjectID string) error {
	const op errs.Op = "accessService.RevokeAccessToDataset"

	ds, err := s.dataProductStorage.GetDataset(ctx, access.DatasetID)
	if err != nil {
		return errs.E(op, err)
	}

	bqds, err := s.bigQueryStorage.GetBigqueryDatasource(ctx, access.DatasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	subjectParts := strings.Split(access.Subject, ":")
	if len(subjectParts) != 2 {
		return errs.E(errs.InvalidRequest, service.CodeUnexpectedSubjectFormat, op, fmt.Errorf("subject is not in the correct format"))
	}

	subjectWithoutType := subjectParts[1]

	if len(bqds.PseudoColumns) > 0 {
		joinableViews, err := s.joinableViewStorage.GetJoinableViewsForReferenceAndUser(ctx, subjectWithoutType, ds.ID)
		if err != nil {
			return errs.E(op, err)
		}

		for _, jv := range joinableViews {
			// FIXME: this is a bit of a hack, we should probably have a better way to get the joinable view name
			joinableViewName := makeJoinableViewName(bqds.ProjectID, bqds.Table)
			if err := s.bigQueryAPI.Revoke(ctx, gcpProjectID, jv.Dataset, joinableViewName, access.Subject); err != nil {
				return errs.E(op, err)
			}
		}
	}

	if err := s.bigQueryAPI.Revoke(ctx, bqds.ProjectID, bqds.Dataset, bqds.Table, access.Subject); err != nil {
		return errs.E(op, err)
	}

	if err := s.accessStorage.RevokeAccessToDataset(ctx, access.ID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) EnsureUserIsAuthorizedToRevokeAccess(ctx context.Context, user *service.User, access *service.Access) error {
	const op errs.Op = "accessService.EnsureUserIsAuthorizedToRevokeAccess"

	err := s.dataProductService.EnsureUserIsOwner(ctx, user, access.DatasetID)
	if err == nil {
		return nil
	}

	if user.Email == access.Owner {
		return nil
	}

	err = ensureUserInGroup(user, access.Owner)
	if err == nil {
		return nil
	}

	return errs.E(errs.Unauthorized, service.CodeWrongTeam, op, errs.UserName(user.Email), fmt.Errorf("user not authorized to revoke access for subject %v", access.Subject))
}

func (s *accessService) EnsureUserIsAuthorizedToApproveRequest(ctx context.Context, user *service.User, accessID uuid.UUID) error {
	const op errs.Op = "accessService.EnsureUserIsAuthorizedToApproveRequest"
	datasetID, err := s.accessStorage.GetDatasetIDFromAccessRequest(ctx, accessID)
	if err != nil {
		return errs.E(op, err)
	}

	return s.EnsureUserIsDatasetOwner(ctx, user, datasetID)
}

func (s *accessService) EnsureUserIsDatasetOwner(ctx context.Context, user *service.User, datasetID uuid.UUID) error {
	const op errs.Op = "accessService.EnsureUserIsDatasetOwner"

	err := s.dataProductService.EnsureUserIsOwner(ctx, user, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func makeJoinableViewName(projectID, tableID string) string {
	// datasetID will always be same markedsplassen dataset id
	return fmt.Sprintf("%v_%v", projectID, tableID)
}

func (s *accessService) GetAccessToDataset(ctx context.Context, accessID uuid.UUID) (*service.Access, error) {
	const op errs.Op = "accessService.GetAccessToDataset"

	access, err := s.accessStorage.GetAccessToDataset(ctx, accessID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return access, nil
}

// nolint: gocyclo
func (s *accessService) GrantBigQueryAccessToDataset(ctx context.Context, user *service.User, input service.GrantAccessData, gcpProjectID string) error {
	const op errs.Op = "accessService.GrantBigQueryAccessToDataset"

	identity := newResolvedAccessIdentity(user, accessTarget{
		input.Subject,
		input.SubjectType,
		input.Owner,
	})

	ds, err := s.dataProductStorage.GetDataset(ctx, input.DatasetID)
	if err != nil {
		return errs.E(op, err)
	}

	if err = s.validateAccessGrant(ds, identity.Subject); err != nil {
		return errs.E(op, err)
	}

	err = s.grantBigQueryAccess(ctx, identity, ds.ID, gcpProjectID)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.accessStorage.GrantAccessToDatasetAndRenew(ctx, input.DatasetID, input.Expires, identity.SubjectWithType(), identity.Owner, user.Email, service.AccessPlatformBigQuery)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) grantBigQueryAccess(ctx context.Context, identity resolvedAccessIdentity, datasetID uuid.UUID, gcpProjectID string) error {
	const op errs.Op = "accessService.grantBigQueryAccess"
	bqds, err := s.bigQueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	if len(bqds.PseudoColumns) > 0 {
		joinableViews, err := s.joinableViewStorage.GetJoinableViewsForReferenceAndUser(ctx, identity.Subject, datasetID)
		if err != nil {
			return errs.E(op, err)
		}

		for _, jv := range joinableViews {
			joinableViewName := makeJoinableViewName(bqds.ProjectID, bqds.Table)
			if err := s.bigQueryAPI.Grant(ctx, gcpProjectID, jv.Dataset, joinableViewName, identity.SubjectWithType()); err != nil {
				return errs.E(op, err)
			}
		}
	}

	if err := s.bigQueryAPI.Grant(ctx, bqds.ProjectID, bqds.Dataset, bqds.Table, identity.SubjectWithType()); err != nil {
		return errs.E(op, err)
	}
	return nil
}

func (s *accessService) GrantMetabaseAccessToDataset(ctx context.Context, user *service.User, input service.GrantAccessData) error {
	const op errs.Op = "accessService.GrantMetabaseAccessToDataset"

	identity := newResolvedAccessIdentity(user, accessTarget{
		input.Subject,
		input.SubjectType,
		input.Owner,
	})

	ds, err := s.dataProductStorage.GetDataset(ctx, input.DatasetID)
	if err != nil {
		return errs.E(op, err)
	}

	if err = s.validateAccessGrant(ds, identity.Subject); err != nil {
		return errs.E(op, err)
	}

	err = s.metabaseService.GrantMetabaseAccessRestricted(ctx, ds.ID, identity.Subject, identity.SubjectType)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.accessStorage.GrantAccessToDatasetAndRenew(ctx, input.DatasetID, input.Expires, identity.SubjectWithType(), identity.Owner, user.Email, service.AccessPlatformMetabase)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) RevokeMetabaseAllUsersAccessToDataset(ctx context.Context, access *service.Access) error {
	const op errs.Op = "accessService.RevokeMetabaseAllUsersAccessToDataset"

	if err := s.metabaseService.RevokeMetabaseAccessAllUsers(ctx, access.DatasetID); err != nil {
		return errs.E(op, err)
	}

	if err := s.accessStorage.RevokeAccessToDataset(ctx, access.ID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) RevokeMetabaseRestrictedAccessToDataset(ctx context.Context, access *service.Access) error {
	const op errs.Op = "accessService.RevokeMetabaseRestrictedAccessToDataset"

	subjectParts := strings.Split(access.Subject, ":")
	subjectType := subjectParts[0]
	subject := subjectParts[1]

	err := s.metabaseService.RevokeMetabaseAccessRestricted(ctx, access.DatasetID, subject, subjectType)
	if err != nil {
		return errs.E(op, err)
	}

	if err := s.accessStorage.RevokeAccessToDataset(ctx, access.ID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *accessService) validateAccessGrant(ds *service.Dataset, subject string) error {
	if ds.Pii == "sensitive" && subject == s.allUsersEmail {
		return errs.E(errs.InvalidRequest, service.CodeOpeningDatasetWithPiiTags, fmt.Errorf("datasett som inneholder personopplysninger kan ikke gjøres tilgjengelig for alle interne brukere"))
	}
	return nil
}

func ensureOwner(user *service.User, owner string) error {
	const op errs.Op = "ensureOwner"

	if user != nil && (user.GoogleGroups.Contains(owner) || owner == user.Email) {
		return nil
	}

	return errs.E(errs.Unauthorized, op, errs.UserName(user.Email), fmt.Errorf("user is not owner"))
}

type resolvedAccessIdentity struct {
	Subject          string
	SubjectType      string
	Owner            string
	OwnerSubjectType string
}

func (c *resolvedAccessIdentity) SubjectWithType() string {
	return c.SubjectType + ":" + c.Subject
}

func (c *resolvedAccessIdentity) OwnerWithType() string {
	return c.OwnerSubjectType + ":" + c.Owner
}

type accessTarget struct {
	Subject     *string
	SubjectType *string
	Owner       *string
}

func newResolvedAccessIdentity(user *service.User, target accessTarget) resolvedAccessIdentity {
	subject := user.Email
	if target.Subject != nil {
		subject = *target.Subject
	}
	subjectType := service.SubjectTypeUser
	if target.SubjectType != nil {
		subjectType = *target.SubjectType
	}
	owner := subject
	if target.Owner != nil && *target.SubjectType == service.SubjectTypeServiceAccount {
		owner = *target.Owner
	}

	ownerSubjectType := subjectType
	if subjectType == service.SubjectTypeServiceAccount {
		if target.Owner != nil {
			ownerSubjectType = service.SubjectTypeGroup
		} else {
			owner = service.SubjectTypeUser
		}
	}

	return resolvedAccessIdentity{
		Subject:          subject,
		SubjectType:      subjectType,
		Owner:            owner,
		OwnerSubjectType: ownerSubjectType,
	}
}

func NewAccessService(
	dataCatalogueURL string,
	allUsersEmail string,
	slackapi service.SlackAPI,
	pollyStorage service.PollyStorage,
	accessStorage service.AccessStorage,
	dataProductStorage service.DataProductsStorage,
	bigQueryStorage service.BigQueryStorage,
	joinableViewStorage service.JoinableViewsStorage,
	bigQueryAPI service.BigQueryAPI,
	dataproductService service.DataProductsService,
	metabaseService service.MetabaseService,
) *accessService {
	return &accessService{
		dataCatalogueURL:    dataCatalogueURL,
		allUsersEmail:       allUsersEmail,
		slackapi:            slackapi,
		pollyStorage:        pollyStorage,
		accessStorage:       accessStorage,
		dataProductStorage:  dataProductStorage,
		bigQueryStorage:     bigQueryStorage,
		joinableViewStorage: joinableViewStorage,
		bigQueryAPI:         bigQueryAPI,
		dataProductService:  dataproductService,
		metabaseService:     metabaseService,
	}
}
