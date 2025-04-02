package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/rs/zerolog"
	"net/http"
	"riverqueue.com/riverui"
)

type RiverHandler struct {
	server *riverui.Server
	logger zerolog.Logger
}

func (h *RiverHandler) Handle(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	user := auth.GetUser(r.Context())

	if !user.IsAdmin() {
		errs.HTTPErrorResponse(w, h.logger, errs.E(errs.Unauthorized, fmt.Errorf("not an administrator")), requestID)

		return
	}

	h.server.ServeHTTP(w, r)
}

func NewRiverHandler(server *riverui.Server, logger zerolog.Logger) *RiverHandler {
	return &RiverHandler{
		server: server,
		logger: logger,
	}
}
