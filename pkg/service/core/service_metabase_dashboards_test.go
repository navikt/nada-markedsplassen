package core

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type metabaseURLCase struct {
	name string
	path string
	host string
	link string
	id   int32
	kind errs.Kind
}

// Paths use the requested resource type so both wrappers exercise identical rules.
func metabaseURLCases(resource string) []metabaseURLCase {
	base := "https://metabase.example.com"
	path := "/" + resource + "/"
	other := "document"
	if resource == other {
		other = "dashboard"
	}
	return []metabaseURLCase{
		{name: "numeric ID", path: path + "1", id: 1},
		{name: "slug", path: path + "1-example-slug", id: 1},
		{name: "leading zeros", path: path + "0001-slug", id: 1},
		{name: "maximum ID", path: path + "2147483647", id: 2147483647},
		{name: "query and fragment", path: path + "1-slug?return=/other/99&value=%2F#section", id: 1},
		{name: "whitespace", link: " " + base + path + "1 ", id: 1},
		{name: "host case", link: "https://METABASE.EXAMPLE.COM" + path + "1", id: 1},
		{name: "configured trailing slash", host: base + "/", path: path + "1", id: 1},
		{name: "configured base path", host: base + "/metabase", link: base + "/metabase" + path + "1-slug", id: 1},
		{name: "configured base path trailing slash", host: base + "/metabase/", link: base + "/metabase" + path + "1", id: 1},
		{name: "localhost", host: "http://localhost:8083", link: "http://localhost:8083" + path + "1", id: 1},
		{name: "localhost on another instance", link: "https://localhost" + path + "1", kind: errs.InvalidRequest},
		{name: "localhost on another port", host: "http://localhost:8083", link: "http://localhost:8084" + path + "1", kind: errs.InvalidRequest},
		{name: "other resource", path: "/" + other + "/1", kind: errs.InvalidRequest},
		{name: "public URL", path: "/public" + path + "1", kind: errs.InvalidRequest},
		{name: "missing ID", path: path, kind: errs.InvalidRequest},
		{name: "zero", path: path + "0", kind: errs.InvalidRequest},
		{name: "negative", path: path + "-1", kind: errs.InvalidRequest},
		{name: "signed number", path: path + "+1", kind: errs.InvalidRequest},
		{name: "overflow", path: path + "2147483648", kind: errs.InvalidRequest},
		{name: "unicode digits", path: path + "１", kind: errs.InvalidRequest},
		{name: "ID suffix", path: path + "1abc", kind: errs.InvalidRequest},
		{name: "empty slug", path: path + "1-", kind: errs.InvalidRequest},
		{name: "trailing slash", path: path + "1/", kind: errs.InvalidRequest},
		{name: "extra segment", path: path + "1/extra", kind: errs.InvalidRequest},
		{name: "dot segments", path: path + "../" + resource + "/1", kind: errs.InvalidRequest},
		{name: "encoded dot segments", path: path + "%2e%2e/1", kind: errs.InvalidRequest},
		{name: "encoded slash", path: path + "1%2Fextra", kind: errs.InvalidRequest},
		{name: "encoded backslash", path: path + "1%5Cextra", kind: errs.InvalidRequest},
		{name: "wrong host", link: "https://other.example.com" + path + "1", kind: errs.InvalidRequest},
		{name: "host suffix", link: "https://metabase.example.com.evil.example" + path + "1", kind: errs.InvalidRequest},
		{name: "wrong scheme", link: "http://metabase.example.com" + path + "1", kind: errs.InvalidRequest},
		{name: "unsupported scheme", link: "ftp://metabase.example.com" + path + "1", kind: errs.InvalidRequest},
		{name: "wrong port", link: "https://metabase.example.com:8443" + path + "1", kind: errs.InvalidRequest},
		{name: "explicit default port", link: "https://metabase.example.com:443" + path + "1", kind: errs.InvalidRequest},
		{name: "userinfo", link: "https://user:password@metabase.example.com" + path + "1", kind: errs.InvalidRequest},
		{name: "empty userinfo", link: "https://@metabase.example.com" + path + "1", kind: errs.InvalidRequest},
		{name: "opaque URL", link: "https:metabase.example.com" + path + "1", kind: errs.InvalidRequest},
		{name: "relative URL", link: path + "1", kind: errs.InvalidRequest},
		{name: "invalid escape", link: base + path + "%zz", kind: errs.InvalidRequest},
		{name: "missing base path", host: base + "/metabase", link: base + path + "1", kind: errs.InvalidRequest},
		{name: "base path prefix collision", host: base + "/metabase", link: base + "/metabase-other" + path + "1", kind: errs.InvalidRequest},
		{name: "empty configuration", host: " ", path: path + "1", kind: errs.Internal},
		{name: "unsupported configuration", host: "ftp://metabase.example.com", path: path + "1", kind: errs.Internal},
		{name: "configuration with userinfo", host: "https://user@metabase.example.com", path: path + "1", kind: errs.Internal},
	}
}

func (tt metabaseURLCase) values() (string, string) {
	host, link := tt.host, tt.link
	if host == "" {
		host = "https://metabase.example.com"
	}
	if link == "" {
		link = "https://metabase.example.com" + tt.path
	}
	return link, host
}

func TestMetabaseResourceURLValidation(t *testing.T) {
	for _, resource := range []struct {
		name  string
		parse func(string, string) (int32, error)
		code  errs.Code
		op    errs.Op
	}{
		{"dashboard", parseMetabaseDashboardID, service.CodeMetabaseDashboardInvalidLink, "parseMetabaseDashboardID"},
	} {
		t.Run(resource.name, func(t *testing.T) {
			for _, tt := range metabaseURLCases(resource.name) {
				t.Run(tt.name, func(t *testing.T) {
					link, host := tt.values()
					id, err := resource.parse(link, host)
					if tt.kind == errs.Other {
						if err != nil || id != tt.id {
							t.Fatalf("id=%d err=%v, want id=%d", id, err, tt.id)
						}
						return
					}
					if id != 0 || !errs.KindIs(tt.kind, err) {
						t.Fatalf("id=%d err=%v, want kind=%v", id, err, tt.kind)
					}
					var detail *errs.Error
					if !errors.As(err, &detail) {
						t.Fatal("expected structured error")
					}
					wantCode, wantMessage := resource.code, "invalid Metabase "+resource.name+" URL"
					if tt.kind == errs.Internal {
						wantCode, wantMessage = service.CodeMetabase, "invalid configured Metabase URL"
					}
					if detail.Code != wantCode || detail.Op != resource.op || err.Error() != wantMessage {
						t.Fatalf("unexpected error contract: code=%s op=%s message=%s", detail.Code, detail.Op, err.Error())
					}
				})
			}
		})
	}
}

type dashboardURLTestAPI struct {
	service.MetabaseAPI
	getIDs     []string
	publishIDs []string
	events     *[]string
	publicID   uuid.UUID
}

func (a *dashboardURLTestAPI) GetDashboard(_ context.Context, id string) (*service.MetabaseDashboard, error) {
	a.getIDs = append(a.getIDs, id)
	*a.events = append(*a.events, "get")
	return &service.MetabaseDashboard{Name: "Synthetic dashboard", CollectionID: 0}, nil
}

func (a *dashboardURLTestAPI) CreatePublicDashboardLink(_ context.Context, id string) (uuid.UUID, error) {
	a.publishIDs = append(a.publishIDs, id)
	*a.events = append(*a.events, "publish")
	return a.publicID, nil
}

type dashboardURLTestStorage struct {
	service.MetabaseDashboardStorage
	created *service.NewPublicMetabaseDashboard
	events  *[]string
}

func (s *dashboardURLTestStorage) CreateMetabaseDashboard(_ context.Context, input *service.NewPublicMetabaseDashboard) (*service.PublicMetabaseDashboard, error) {
	s.created = input
	*s.events = append(*s.events, "store")
	return &service.PublicMetabaseDashboard{ID: uuid.New(), MetabaseID: int(input.MetabaseID), Name: input.Name,
		PublicDashboardID: input.PublicDashboardID, Group: input.Input.Group, Keywords: input.Input.Keywords}, nil
}

func TestMetabaseDashboardURLValidationBeforeAPICalls(t *testing.T) {
	for _, tt := range metabaseURLCases("dashboard") {
		t.Run(tt.name, func(t *testing.T) {
			events := []string{}
			api := &dashboardURLTestAPI{events: &events, publicID: uuid.New()}
			storage := &dashboardURLTestStorage{events: &events}
			link, host := tt.values()
			s := NewMetabaseDashboardsService(storage, api, host)
			output, err := s.CreateMetabaseDashboard(context.Background(), &service.User{Email: "developer@example.com"}, service.PublicMetabaseDashboardInput{
				Link: link, Group: "owner@example.com",
			})
			if tt.kind != errs.Other {
				if !errs.KindIs(tt.kind, err) || output != nil {
					t.Fatalf("output=%#v err=%v, want kind=%v", output, err, tt.kind)
				}
				var detail *errs.Error
				if !errors.As(err, &detail) {
					t.Fatal("expected structured error")
				}
				wantCode := service.CodeMetabaseDashboardInvalidLink
				if tt.kind == errs.Internal {
					wantCode = service.CodeMetabase
				}
				if detail.Code != wantCode || detail.User != "" {
					t.Fatalf("unexpected validation error: %#v", detail)
				}
				if len(events) != 0 || len(api.getIDs) != 0 || len(api.publishIDs) != 0 || storage.created != nil {
					t.Fatalf("invalid URL caused side effects: events=%v get=%v publish=%v", events, api.getIDs, api.publishIDs)
				}
				if strings.Contains(err.Error(), link) {
					t.Fatal("validation error contains the supplied URL")
				}
				return
			}
			if err != nil || output == nil {
				t.Fatalf("output=%#v err=%v", output, err)
			}
			wantID := strconv.Itoa(int(tt.id))
			if !reflect.DeepEqual(api.getIDs, []string{wantID}) || !reflect.DeepEqual(api.publishIDs, []string{wantID}) {
				t.Fatalf("non-normalized ID sent upstream: get=%v publish=%v", api.getIDs, api.publishIDs)
			}
			if storage.created.MetabaseID != tt.id || storage.created.PublicDashboardID != api.publicID {
				t.Fatalf("incorrect stored IDs: %#v", storage.created)
			}
			if !reflect.DeepEqual(events, []string{"get", "publish", "store"}) {
				t.Fatalf("unexpected publishing sequence: %v", events)
			}
		})
	}
}
