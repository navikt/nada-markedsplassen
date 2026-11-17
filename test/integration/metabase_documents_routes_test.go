package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/rs/zerolog"
)

type documentRouteService struct {
	service.MetabaseDocumentsService
	calls  int
	err    error
	update service.PublicMetabaseDocumentEditInput
}

func (s *documentRouteService) CreateMetabaseDocument(context.Context, *service.User, service.PublicMetabaseDocumentInput) (*service.PublicMetabaseDocumentOutput, error) {
	s.calls++
	return &service.PublicMetabaseDocumentOutput{Name: "Synthetic document"}, s.err
}
func (s *documentRouteService) GetMetabaseDocument(context.Context, uuid.UUID) (*service.PublicMetabaseDocumentOutput, error) {
	s.calls++
	return &service.PublicMetabaseDocumentOutput{Name: "Synthetic document"}, s.err
}
func (s *documentRouteService) UpdateMetabaseDocument(_ context.Context, _ *service.User, _ uuid.UUID, input service.PublicMetabaseDocumentEditInput) (*service.PublicMetabaseDocumentOutput, error) {
	s.calls++
	s.update = input
	return &service.PublicMetabaseDocumentOutput{Name: "Synthetic document"}, s.err
}
func (s *documentRouteService) DeleteMetabaseDocument(context.Context, *service.User, uuid.UUID) error {
	s.calls++
	return s.err
}

func TestMetabaseDocumentRoutes(t *testing.T) {
	id := uuid.New().String()
	for _, route := range []struct{ method, path string }{
		{"POST", "/api/metabase-documents"}, {"POST", "/api/metabase-documents/"},
		{"GET", "/api/metabase-documents/" + id}, {"PUT", "/api/metabase-documents/" + id}, {"DELETE", "/api/metabase-documents/" + id},
	} {
		for _, loggedIn := range []bool{false, true} {
			t.Run(route.method+route.path, func(t *testing.T) {
				s := &documentRouteService{}
				middlewareCalls := 0
				middleware := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						middlewareCalls++
						if !loggedIn {
							w.WriteHeader(http.StatusUnauthorized)
							return
						}
						// Login without group membership must still permit public metadata GET.
						r = r.WithContext(auth.SetUser(r.Context(), &service.User{Email: "developer@example.com"}))
						next.ServeHTTP(w, r)
					})
				}
				router := chi.NewRouter()
				routes.NewMetabaseDocumentRouter(routes.NewMetabaseDocumentEndpoints(zerolog.Nop(), handlers.NewMetabaseDocumentsHandler(s)), middleware)(router)
				request := httptest.NewRequest(route.method, route.path, strings.NewReader(`{}`))
				request.Header.Set("Content-Type", "application/json")
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				if middlewareCalls != 1 {
					t.Fatalf("auth middleware calls=%d", middlewareCalls)
				}
				if !loggedIn {
					if response.Code != 401 || s.calls != 0 {
						t.Fatalf("status=%d calls=%d", response.Code, s.calls)
					}
				} else {
					wantStatus := http.StatusOK
					if route.method == http.MethodDelete {
						wantStatus = http.StatusNoContent
					}
					if response.Code != wantStatus || s.calls != 1 {
						t.Fatalf("status=%d body=%s calls=%d", response.Code, response.Body, s.calls)
					}
				}
			})
		}
	}
}

// Handlers that act on behalf of a user must reject requests without one, even
// if the auth middleware is bypassed. GET only reads public metadata and relies
// on the middleware, like the dashboard handler.
func TestMetabaseDocumentHandlersRequireUser(t *testing.T) {
	id := uuid.New().String()
	for _, method := range []string{"POST", "PUT", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			s := &documentRouteService{}
			router := chi.NewRouter()
			passThrough := func(next http.Handler) http.Handler { return next }
			routes.NewMetabaseDocumentRouter(routes.NewMetabaseDocumentEndpoints(zerolog.Nop(), handlers.NewMetabaseDocumentsHandler(s)), passThrough)(router)
			path := "/api/metabase-documents/" + id
			if method == "POST" {
				path = "/api/metabase-documents"
			}
			req := httptest.NewRequest(method, path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != 401 || s.calls != 0 {
				t.Fatalf("status=%d calls=%d", rec.Code, s.calls)
			}
		})
	}
}

func TestMetabaseDocumentRouteInputAndErrors(t *testing.T) {
	id := uuid.New().String()
	for _, tt := range []struct {
		name, method, path, body string
		err                      error
		status                   int
	}{
		{"invalid ID", "GET", "/api/metabase-documents/not-a-uuid", `{}`, nil, 400},
		{"invalid JSON", "POST", "/api/metabase-documents", `{`, nil, 400},
		{"forbidden", "DELETE", "/api/metabase-documents/" + id, `{}`, errs.E(errs.Unauthorized, "not authorized"), 403},
		{"not found", "GET", "/api/metabase-documents/" + id, `{}`, errs.E(errs.NotExist, "not found"), 404},
		{"group cannot change", "PUT", "/api/metabase-documents/" + id, `{"group":"other@example.com","description":"updated"}`, nil, 200},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := &documentRouteService{err: tt.err}
			middleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r.WithContext(auth.SetUser(r.Context(), &service.User{Email: "developer@example.com"})))
				})
			}
			router := chi.NewRouter()
			routes.NewMetabaseDocumentRouter(routes.NewMetabaseDocumentEndpoints(zerolog.Nop(), handlers.NewMetabaseDocumentsHandler(s)), middleware)(router)
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body)
			}
			if tt.name == "group cannot change" && (s.update.Description == nil || *s.update.Description != "updated") {
				t.Fatalf("input=%#v", s.update)
			}
		})
	}
}
