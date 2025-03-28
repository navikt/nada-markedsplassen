package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type AccessQueries interface {
	ListAccessRequestsForOwner(ctx context.Context, owner []string) ([]gensql.DatasetAccessRequest, error)
	ListUnrevokedExpiredAccessEntries(ctx context.Context) ([]gensql.DatasetAccessView, error)
	ListActiveAccessToDataset(ctx context.Context, datasetID uuid.UUID) ([]gensql.DatasetAccessView, error)
	ListAccessRequestsForDataset(ctx context.Context, datasetID uuid.UUID) ([]gensql.DatasetAccessRequest, error)
	CreateAccessRequestForDataset(ctx context.Context, params gensql.CreateAccessRequestForDatasetParams) (gensql.DatasetAccessRequest, error)
	GetAccessRequest(ctx context.Context, id uuid.UUID) (gensql.DatasetAccessRequest, error)
	DeleteAccessRequest(ctx context.Context, id uuid.UUID) error
	UpdateAccessRequest(ctx context.Context, params gensql.UpdateAccessRequestParams) (gensql.DatasetAccessRequest, error)
	GrantAccessToDataset(ctx context.Context, params gensql.GrantAccessToDatasetParams) error
	ApproveAccessRequest(ctx context.Context, params gensql.ApproveAccessRequestParams) error
	GetActiveAccessToDatasetForSubject(ctx context.Context, params gensql.GetActiveAccessToDatasetForSubjectParams) (gensql.DatasetAccessView, error)
	RevokeAccessToDataset(ctx context.Context, id uuid.UUID) error
	DenyAccessRequest(ctx context.Context, params gensql.DenyAccessRequestParams) error
	GetAccessToDataset(ctx context.Context, id uuid.UUID) (gensql.DatasetAccessView, error)
}

var _ service.AccessStorage = &accessStorage{}

type AccessQueriesWithTxFn func() (AccessQueries, database.Transacter, error)

type accessStorage struct {
	queries  AccessQueries
	withTxFn AccessQueriesWithTxFn
}

func (s *accessStorage) ListAccessRequestsForOwner(ctx context.Context, owner []string) ([]service.AccessRequest, error) {
	const op errs.Op = "accessStorage.ListAccessRequestsForOwner"

	raw, err := s.queries.ListAccessRequestsForOwner(ctx, owner)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err, service.ParamOwner)
	}

	accessRequests, err := From(DatasetAccessRequests(raw))
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamAccessRequest)
	}

	return accessRequests, nil
}

func (s *accessStorage) GetUnrevokedExpiredAccess(ctx context.Context) ([]*service.Access, error) {
	const op errs.Op = "accessStorage.GetUnrevokedExpiredAccess"

	expired, err := s.queries.ListUnrevokedExpiredAccessEntries(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	var ret []*service.Access
	for _, e := range expired {
		ar, _ := From(DatasetAccess(e))
		ret = append(ret, ar)
	}

	return ret, nil
}

func (s *accessStorage) ListActiveAccessToDataset(ctx context.Context, datasetID uuid.UUID) ([]*service.Access, error) {
	const op errs.Op = "accessStorage.ListActiveAccessToDataset"

	access, err := s.queries.ListActiveAccessToDataset(ctx, datasetID)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op)
	}

	var ret []*service.Access
	for _, e := range access {
		ar, _ := From(DatasetAccess(e))
		ret = append(ret, ar)
	}

	return ret, nil
}

func (s *accessStorage) ListAccessRequestsForDataset(ctx context.Context, datasetID uuid.UUID) ([]service.AccessRequest, error) {
	const op errs.Op = "accessStorage.ListAccessRequestsForDataset"

	raw, err := s.queries.ListAccessRequestsForDataset(ctx, datasetID)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	accessRequests, err := DatasetAccessRequests(raw).To()
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamAccessRequest)
	}

	return accessRequests, nil
}

func (s *accessStorage) CreateAccessRequestForDataset(ctx context.Context, datasetID uuid.UUID, pollyDocumentationID uuid.NullUUID, subject, owner string, expires *time.Time) (*service.AccessRequest, error) {
	const op errs.Op = "accessStorage.CreateAccessRequestForDataset"

	raw, err := s.queries.CreateAccessRequestForDataset(ctx, gensql.CreateAccessRequestForDatasetParams{
		DatasetID:            datasetID,
		Subject:              emailOfSubjectToLower(subject),
		Owner:                strings.Split(owner, ":")[1],
		Expires:              ptrToNullTime(expires),
		PollyDocumentationID: pollyDocumentationID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, err, service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	ar, err := From(DatasetAccessRequest(raw))
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamAccessRequest)
	}

	return ar, nil
}

func (s *accessStorage) GetAccessRequest(ctx context.Context, accessRequestID uuid.UUID) (*service.AccessRequest, error) {
	const op errs.Op = "accessStorage.GetAccessRequest"

	raw, err := s.queries.GetAccessRequest(ctx, accessRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, err, service.ParamAccessRequest)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	ar, err := From(DatasetAccessRequest(raw))
	if err != nil {
		return nil, errs.E(errs.Internal, service.CodeInternalDecoding, op, err, service.ParamAccessRequest)
	}

	return ar, nil
}

func (s *accessStorage) DeleteAccessRequest(ctx context.Context, accessRequestID uuid.UUID) error {
	const op errs.Op = "accessStorage.DeleteAccessRequest"

	err := s.queries.DeleteAccessRequest(ctx, accessRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *accessStorage) UpdateAccessRequest(ctx context.Context, input service.UpdateAccessRequestDTO) error {
	const op errs.Op = "accessStorage.UpdateAccessRequest"

	var pollyID uuid.NullUUID

	if input.Polly != nil && input.Polly.ID != nil {
		pollyID = uuid.NullUUID{UUID: *input.Polly.ID, Valid: true}
	}

	_, err := s.queries.UpdateAccessRequest(ctx, gensql.UpdateAccessRequestParams{
		Owner:                input.Owner,
		Expires:              ptrToNullTime(input.Expires),
		PollyDocumentationID: pollyID,
		ID:                   input.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.E(errs.NotExist, service.CodeDatabase, op, err, service.ParamAccessRequest)
		}

		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *accessStorage) GrantAccessToDatasetAndApproveRequest(ctx context.Context, user *service.User, datasetID uuid.UUID, subject, accessRequestOwner string, accessRequestID uuid.UUID, expires *time.Time) error {
	const op errs.Op = "accessStorage.GrantAccessToDatasetAndApproveRequest"

	q, tx, err := s.withTxFn()
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}
	defer tx.Rollback()

	err = q.GrantAccessToDataset(ctx, gensql.GrantAccessToDatasetParams{
		DatasetID: datasetID,
		Subject:   subject,
		Granter:   user.Email,
		Owner:     accessRequestOwner,
		Expires:   ptrToNullTime(expires),
		AccessRequestID: uuid.NullUUID{
			UUID:  accessRequestID,
			Valid: true,
		},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.E(errs.NotExist, service.CodeDatabase, op, err)
		}

		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	err = q.ApproveAccessRequest(ctx, gensql.ApproveAccessRequestParams{
		ID:      accessRequestID,
		Granter: sql.NullString{String: user.Email, Valid: true},
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	err = tx.Commit()
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *accessStorage) GrantAccessToDatasetAndRenew(ctx context.Context, datasetID uuid.UUID, expires *time.Time, subject, owner, granter string) error {
	const op errs.Op = "accessStorage.GrantAccessToDatasetAndRenew"

	a, err := s.queries.GetActiveAccessToDatasetForSubject(ctx, gensql.GetActiveAccessToDatasetForSubjectParams{
		DatasetID: datasetID,
		Subject:   subject,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	q, tx, err := s.withTxFn()
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}
	defer tx.Rollback()

	if len(a.AccessSubject) > 0 {
		if err := q.RevokeAccessToDataset(ctx, a.AccessID); err != nil {
			return errs.E(errs.Database, service.CodeDatabase, op, err)
		}
	}

	err = q.GrantAccessToDataset(ctx, gensql.GrantAccessToDatasetParams{
		DatasetID: datasetID,
		Subject:   emailOfSubjectToLower(subject),
		Expires:   ptrToNullTime(expires),
		Owner:     owner,
		Granter:   granter,
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	err = tx.Commit()
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *accessStorage) DenyAccessRequest(ctx context.Context, user *service.User, accessRequestID uuid.UUID, reason *string) error {
	const op errs.Op = "accessStorage.DenyAccessRequest"

	err := s.queries.DenyAccessRequest(ctx, gensql.DenyAccessRequestParams{
		ID:      accessRequestID,
		Granter: sql.NullString{String: user.Email, Valid: true},
		Reason:  ptrToNullString(reason),
	})
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *accessStorage) GetAccessToDataset(ctx context.Context, id uuid.UUID) (*service.Access, error) {
	const op errs.Op = "accessStorage.GetAccessToDataset"

	access, err := s.queries.GetAccessToDataset(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, err, service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return &service.Access{
		ID:        access.AccessID,
		Subject:   access.AccessSubject,
		Granter:   access.AccessGranter,
		Owner:     access.AccessOwner,
		Expires:   nullTimeToPtr(access.AccessExpires),
		Created:   access.AccessCreated,
		Revoked:   nullTimeToPtr(access.AccessRevoked),
		DatasetID: access.AccessDatasetID,
		AccessRequest: &service.AccessRequest{
			ID:          access.AccessRequestID.UUID,
			DatasetID:   access.AccessID,
			Subject:     access.AccessRequestOwner.String,
			SubjectType: strings.Split(access.AccessRequestSubject.String, ":")[0],
			Created:     access.AccessRequestCreated.Time,
			Expires:     nullTimeToPtr(access.AccessRequestExpires),
			Closed:      nullTimeToPtr(access.AccessRequestClosed),
			Granter:     nullStringToPtr(access.AccessRequestGranter),
			Owner:       access.AccessRequestOwner.String,
			Reason:      nullStringToPtr(access.AccessRequestReason),
			Status:      service.AccessRequestStatus(access.AccessRequestStatus.AccessRequestStatusType),
			Polly: &service.Polly{
				ID: access.PollyID.UUID,
				QueryPolly: service.QueryPolly{
					ExternalID: access.PollyExternalID.String,
					Name:       access.PollyName.String,
					URL:        access.PollyUrl.String,
				},
			},
		},
	}, nil
}

func (s *accessStorage) RevokeAccessToDataset(ctx context.Context, id uuid.UUID) error {
	const op errs.Op = "accessStorage.RevokeAccessToDataset"

	err := s.queries.RevokeAccessToDataset(ctx, id)
	if err != nil {
		return errs.E(errs.Database, op, err)
	}

	return nil
}

type DatasetAccess gensql.DatasetAccessView

func (a DatasetAccess) To() (*service.Access, error) {
	return &service.Access{
		ID:        a.AccessID,
		Subject:   a.AccessSubject,
		Granter:   a.AccessGranter,
		Expires:   nullTimeToPtr(a.AccessExpires),
		Created:   a.AccessCreated,
		Revoked:   nullTimeToPtr(a.AccessRevoked),
		DatasetID: a.AccessDatasetID,
		AccessRequest: &service.AccessRequest{
			ID:          a.AccessRequestID.UUID,
			DatasetID:   a.AccessID,
			Subject:     a.AccessRequestOwner.String,
			SubjectType: strings.Split(a.AccessRequestSubject.String, ":")[0],
			Created:     a.AccessRequestCreated.Time,
			Expires:     nullTimeToPtr(a.AccessRequestExpires),
			Closed:      nullTimeToPtr(a.AccessRequestClosed),
			Granter:     nullStringToPtr(a.AccessRequestGranter),
			Owner:       a.AccessRequestOwner.String,
			Reason:      nullStringToPtr(a.AccessRequestReason),
			Status:      service.AccessRequestStatus(a.AccessRequestStatus.AccessRequestStatusType),
			Polly: &service.Polly{
				ID: a.PollyID.UUID,
				QueryPolly: service.QueryPolly{
					ExternalID: a.PollyExternalID.String,
					Name:       a.PollyName.String,
					URL:        a.PollyUrl.String,
				},
			},
		},
	}, nil
}

type DatasetAccessRequest gensql.DatasetAccessRequest

func (d DatasetAccessRequest) To() (*service.AccessRequest, error) {
	const op errs.Op = "DatasetAccessRequest.To"

	splits := strings.Split(d.Subject, ":")
	if len(splits) != 2 {
		return nil, errs.E(errs.Internal, op, fmt.Errorf("%v is not a valid subject, expected [subject_type:subject]", d.Subject))
	}

	subject := splits[1]
	subjectType := splits[0]

	switch subjectType {
	case service.SubjectTypeServiceAccount:
	case service.SubjectTypeUser:
	case service.SubjectTypeGroup:
	default:
		return nil, errs.E(errs.Internal, op, fmt.Errorf("%v is not a valid subject type, expected one of [service_account, user, group]", subjectType))
	}

	status, err := From(AccessRequestStatusType(d.Status))
	if err != nil {
		return nil, errs.E(op, err)
	}

	var polly *service.Polly

	if d.PollyDocumentationID.Valid {
		polly = &service.Polly{
			ID: d.PollyDocumentationID.UUID,
		}
	}

	return &service.AccessRequest{
		ID:          d.ID,
		DatasetID:   d.DatasetID,
		Subject:     subject,
		SubjectType: subjectType,
		Created:     d.Created,
		Status:      status,
		Closed:      nullTimeToPtr(d.Closed),
		Expires:     nullTimeToPtr(d.Expires),
		Granter:     nullStringToPtr(d.Granter),
		Owner:       d.Owner,
		Polly:       polly,
		Reason:      nullStringToPtr(d.Reason),
	}, nil
}

type DatasetAccessRequests []gensql.DatasetAccessRequest

func (d DatasetAccessRequests) To() ([]service.AccessRequest, error) {
	const op errs.Op = "DatasetAccessRequests.To"

	accessRequests := make([]service.AccessRequest, len(d))

	for i, raw := range d {
		ar, err := DatasetAccessRequest(raw).To()
		if err != nil {
			return nil, errs.E(op, err)
		}

		accessRequests[i] = *ar
	}

	return accessRequests, nil
}

type AccessRequestStatusType gensql.AccessRequestStatusType

func (a AccessRequestStatusType) To() (service.AccessRequestStatus, error) {
	const op errs.Op = "accessStorage.accessRequestStatusFromDB"

	switch gensql.AccessRequestStatusType(a) {
	case gensql.AccessRequestStatusTypePending:
		return service.AccessRequestStatusPending, nil
	case gensql.AccessRequestStatusTypeApproved:
		return service.AccessRequestStatusApproved, nil
	case gensql.AccessRequestStatusTypeDenied:
		return service.AccessRequestStatusDenied, nil
	default:
		return "", errs.E(op, fmt.Errorf("unknown access request status %q", a))
	}
}

func NewAccessStorage(queries AccessQueries, fn AccessQueriesWithTxFn) *accessStorage {
	return &accessStorage{
		withTxFn: fn,
		queries:  queries,
	}
}
