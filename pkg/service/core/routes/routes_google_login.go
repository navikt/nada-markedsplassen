package routes

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const CookieNameGoogleOAuth2State = "oauth2-google-state"

var googleOauthConfig oauth2.Config

func NewGoogleLoginRoutes(cfg config.Config) AddRoutesFn {
	googleOauthConfig = oauth2.Config{
		RedirectURL:  cfg.OauthGoogle.RedirectURL,
		ClientID:     cfg.OauthGoogle.ClientID,
		ClientSecret: cfg.OauthGoogle.ClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	return func(router chi.Router) {
		router.Route("/api/googleOauth2", func(r chi.Router) {
			r.HandleFunc("/login", GoogleLoginHandler)
			r.HandleFunc("/callback", GoogleCallbackHandler)
		})
	}
}

func CreateHMAC(m string, k string) string {
	h := hmac.New(sha256.New, []byte(k))
	h.Write([]byte(m))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GoogleLoginHandler redirects the user to Google's OAuth 2.0 server
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	redirectUrl := "/"
	if text, ok := query["redirect"]; ok {
		redirectUrl = text[0]
	}

	state := uuid.New().String() + "," + redirectUrl
	stateHMAC := CreateHMAC(state, googleOauthConfig.ClientSecret)
	http.SetCookie(w, &http.Cookie{
		Name:  CookieNameGoogleOAuth2State,
		Value: stateHMAC,
	})
	url := googleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler handles the callback from Google with the authorization code
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(CookieNameGoogleOAuth2State)
	if err != nil {
		http.Error(w, "State token not found in cookie", http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")

	if CreateHMAC(r.FormValue("state"), googleOauthConfig.ClientSecret) != cookie.Value {
		http.Error(w, "State token mismatch", http.StatusBadRequest)
		return
	}

	redirectUrl := strings.Split(state, ",")[1]

	code := r.FormValue("code")
	_, err = googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}
