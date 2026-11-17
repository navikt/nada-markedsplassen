package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type documentStorageFake struct {
	service.MetabaseDocumentStorage
	record                       *service.PublicMetabaseDocument
	getErr, createErr, deleteErr error
	saved                        *service.NewPublicMetabaseDocument
	updated                      *service.EditPublicMetabaseDocument
	events                       *[]string
}

func (f *documentStorageFake) GetMetabaseDocument(context.Context, uuid.UUID) (*service.PublicMetabaseDocument, error) {
	*f.events = append(*f.events, "get-record")
	return f.record, f.getErr
}

func (f *documentStorageFake) CreateMetabaseDocument(_ context.Context, input *service.NewPublicMetabaseDocument) (*service.PublicMetabaseDocument, error) {
	*f.events = append(*f.events, "create-record")
	f.saved = input
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &service.PublicMetabaseDocument{
		ID: uuid.New(), Name: input.Name, Group: input.Input.Group,
		PublicDocumentID: input.PublicDocumentID, MetabaseID: int(input.MetabaseID),
		CreatedBy: input.CreatorEmail, Keywords: input.Input.Keywords,
	}, nil
}

func (f *documentStorageFake) UpdateMetabaseDocument(_ context.Context, input *service.EditPublicMetabaseDocument) (*service.PublicMetabaseDocument, error) {
	*f.events = append(*f.events, "update-record")
	f.updated = input
	updated := *f.record
	updated.Description = input.Input.Description
	updated.Keywords = input.Input.Keywords
	return &updated, nil
}

func (f *documentStorageFake) DeleteMetabaseDocument(context.Context, uuid.UUID) error {
	*f.events = append(*f.events, "delete-record")
	return f.deleteErr
}

type documentAPIFake struct {
	service.MetabaseAPI
	document    *service.MetabaseDocument
	deleteErr   error
	publicID    uuid.UUID
	permissions *service.MetabaseCollectionPermissions
	members     []service.MetabasePermissionGroupMember
	events      *[]string
	lastID      int
}

func (f *documentAPIFake) GetDocument(_ context.Context, id int) (*service.MetabaseDocument, error) {
	*f.events = append(*f.events, "get-document")
	f.lastID = id
	return f.document, nil
}

func (f *documentAPIFake) CreatePublicDocumentLink(_ context.Context, id int) (uuid.UUID, error) {
	*f.events = append(*f.events, "publish")
	f.lastID = id
	return f.publicID, nil
}

func (f *documentAPIFake) DeletePublicDocumentLink(_ context.Context, id int) error {
	*f.events = append(*f.events, "revoke")
	f.lastID = id
	return f.deleteErr
}

func (f *documentAPIFake) GetCollection(_ context.Context, id int) (*service.MetabaseCollection, error) {
	*f.events = append(*f.events, "get-collection")
	return &service.MetabaseCollection{ID: id, Location: "/9/"}, nil
}

func (f *documentAPIFake) GetCollectionPermissions(context.Context) (*service.MetabaseCollectionPermissions, error) {
	*f.events = append(*f.events, "get-permissions")
	return f.permissions, nil
}

func (f *documentAPIFake) GetPermissionGroup(context.Context, int) ([]service.MetabasePermissionGroupMember, error) {
	*f.events = append(*f.events, "get-members")
	return f.members, nil
}

func documentFixture() (*metabaseDocumentsService, *documentStorageFake, *documentAPIFake, *service.User, *[]string) {
	events := []string{}
	storage := &documentStorageFake{events: &events, record: &service.PublicMetabaseDocument{
		ID: uuid.New(), Name: "Synthetic document", Group: "owner@example.com", MetabaseID: 42, PublicDocumentID: uuid.New(),
	}}
	api := &documentAPIFake{
		events:      &events,
		document:    &service.MetabaseDocument{ID: 42, Name: "Synthetic document"},
		publicID:    uuid.New(),
		permissions: &service.MetabaseCollectionPermissions{Groups: map[string]map[string]string{}},
	}
	user := &service.User{Email: "developer@example.com", GoogleGroups: service.GoogleGroups{{Email: "owner@example.com"}}}
	return NewMetabaseDocumentsService(storage, api, "https://metabase.example.com"), storage, api, user, &events
}

func documentInput() service.PublicMetabaseDocumentInput {
	return service.PublicMetabaseDocumentInput{Link: "https://metabase.example.com/document/42-synthetic", Group: "owner@example.com"}
}

func assertDocumentError(t *testing.T, err error, kind errs.Kind, code errs.Code) {
	t.Helper()
	if !errs.KindIs(kind, err) {
		t.Fatalf("error=%v, want kind %v", err, kind)
	}
	var detail *errs.Error
	if code != "" && (!errors.As(err, &detail) || detail.Code != code) {
		t.Fatalf("error=%#v, want code %s", err, code)
	}
}

func assertNoDocumentEvent(t *testing.T, events []string, name string) {
	t.Helper()
	for _, event := range events {
		if event == name {
			t.Fatalf("unexpected %s in %v", name, events)
		}
	}
}

func TestMetabaseDocumentCreate(t *testing.T) {
	// Metabase represents the root collection as null or 0 for documents.
	for name, root := range map[string]*int{"null root": nil, "zero root": new(int)} {
		t.Run(name, func(t *testing.T) {
			s, storage, api, user, events := documentFixture()
			api.document.CollectionID = root

			output, err := s.CreateMetabaseDocument(context.Background(), user, documentInput())
			if err != nil {
				t.Fatal(err)
			}
			if storage.saved.Input.Keywords == nil || api.lastID != 42 {
				t.Fatalf("invalid saved input: %#v", storage.saved)
			}
			if output.Link != "https://metabase.example.com/public/document/"+api.publicID.String() {
				t.Fatal(output.Link)
			}
			if !reflect.DeepEqual(*events, []string{"get-document", "publish", "create-record"}) {
				t.Fatal(*events)
			}
		})
	}
}

func TestMetabaseDocumentCreateRejectsBeforePublishing(t *testing.T) {
	t.Run("wrong owner group", func(t *testing.T) {
		s, _, _, user, events := documentFixture()
		input := documentInput()
		input.Group = "other@example.com"
		_, err := s.CreateMetabaseDocument(context.Background(), user, input)
		assertDocumentError(t, err, errs.Unauthorized, service.CodeWrongOwner)
		if len(*events) != 0 {
			t.Fatal(*events)
		}
	})
	t.Run("invalid link", func(t *testing.T) {
		s, _, _, user, events := documentFixture()
		input := documentInput()
		input.Link = "https://evil.example/document/42"
		_, err := s.CreateMetabaseDocument(context.Background(), user, input)
		assertDocumentError(t, err, errs.InvalidRequest, service.CodeMetabaseDocumentInvalidLink)
		if len(*events) != 0 {
			t.Fatal(*events)
		}
	})
}

func TestMetabaseDocumentCollectionPermissions(t *testing.T) {
	tests := []struct {
		name    string
		groups  map[string]map[string]string
		members bool
		allowed bool
	}{
		{name: "all users write", groups: map[string]map[string]string{"1": {"7": "write", "9": "write"}}, allowed: true},
		{name: "read only", groups: map[string]map[string]string{"1": {"7": "read", "9": "write"}}},
		{name: "no access", groups: map[string]map[string]string{"1": {"7": "none", "9": "write"}}},
		{name: "parent denies", groups: map[string]map[string]string{"1": {"7": "write", "9": "read"}}},
		{name: "member of write group", groups: map[string]map[string]string{"2": {"7": "write", "9": "write"}}, members: true, allowed: true},
		{name: "not member of write group", groups: map[string]map[string]string{"2": {"7": "write", "9": "write"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _, api, user, events := documentFixture()
			collectionID := 7
			api.document.CollectionID = &collectionID
			api.permissions.Groups = tt.groups
			if tt.members {
				api.members = []service.MetabasePermissionGroupMember{{Email: user.Email}}
			}

			_, err := s.CreateMetabaseDocument(context.Background(), user, documentInput())
			if tt.allowed {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			assertDocumentError(t, err, errs.Unauthorized, service.CodeInsufficientPrivileges)
			assertNoDocumentEvent(t, *events, "publish")
		})
	}
}

func TestMetabaseDocumentUpdate(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		s, storage, _, user, _ := documentFixture()
		description := "Updated synthetic description"
		output, err := s.UpdateMetabaseDocument(context.Background(), user, storage.record.ID, service.PublicMetabaseDocumentEditInput{Description: &description})
		if err != nil {
			t.Fatal(err)
		}
		if output.Name != storage.record.Name || *output.Description != description || storage.updated.Input.Keywords == nil {
			t.Fatalf("%#v", output)
		}
	})
	t.Run("wrong owner", func(t *testing.T) {
		s, storage, _, user, events := documentFixture()
		user.GoogleGroups = nil
		_, err := s.UpdateMetabaseDocument(context.Background(), user, storage.record.ID, service.PublicMetabaseDocumentEditInput{})
		assertDocumentError(t, err, errs.Unauthorized, "")
		assertNoDocumentEvent(t, *events, "update-record")
	})
}

func TestMetabaseDocumentDelete(t *testing.T) {
	t.Run("revokes before deleting record", func(t *testing.T) {
		s, storage, _, user, events := documentFixture()
		if err := s.DeleteMetabaseDocument(context.Background(), user, storage.record.ID); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(*events, []string{"get-record", "revoke", "delete-record"}) {
			t.Fatal(*events)
		}
	})
	t.Run("keeps record when revocation fails", func(t *testing.T) {
		s, storage, api, user, events := documentFixture()
		api.deleteErr = errs.E(errs.IO, "metabase unavailable")
		err := s.DeleteMetabaseDocument(context.Background(), user, storage.record.ID)
		assertDocumentError(t, err, errs.IO, "")
		assertNoDocumentEvent(t, *events, "delete-record")
	})
	t.Run("wrong owner", func(t *testing.T) {
		s, storage, _, user, events := documentFixture()
		user.GoogleGroups = nil
		err := s.DeleteMetabaseDocument(context.Background(), user, storage.record.ID)
		assertDocumentError(t, err, errs.Unauthorized, "")
		assertNoDocumentEvent(t, *events, "revoke")
		assertNoDocumentEvent(t, *events, "delete-record")
	})
}

func TestPublicMetabaseDocumentToOutput(t *testing.T) {
	created := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	document := &service.PublicMetabaseDocument{
		ID:               uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Name:             "Example document",
		Group:            "team@example.no",
		PublicDocumentID: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		MetabaseID:       42,
		CreatedBy:        "creator@example.no",
		Created:          created,
		LastModified:     created.Add(time.Hour),
		Keywords:         []string{"report"},
	}

	got := PublicMetabaseDocumentToOutput(document, "https://metabase.example.no")
	if got.Link != "https://metabase.example.no/public/document/33333333-3333-4333-8333-333333333333" {
		t.Fatal(got.Link)
	}
	if got.MetabaseLink != "https://metabase.example.no/document/42" {
		t.Fatal(got.MetabaseLink)
	}
}
