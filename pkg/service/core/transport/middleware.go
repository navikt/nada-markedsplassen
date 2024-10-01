package transport

import (
	"net/http"

	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/service"
)

type MiddlewareHandler func(http.Handler) http.Handler

func UserInfo(withinAnyGroup ...string) MiddlewareHandler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := auth.GetUser(r.Context())
			if user == nil {
				http.Error(w, `{"error": "No user in context"}`, http.StatusUnauthorized)
				return
			}

			if len(withinAnyGroup) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			for _, group := range withinAnyGroup {
				if containsGroup(user.GoogleGroups, group) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error": "User not in authorized group"}`, http.StatusUnauthorized)
		})
	}
}

func containsGroup(groups service.Groups, groupEmail string) bool {
	for _, g := range groups {
		if g.Email == groupEmail {
			return true
		}
	}

	return false
}
