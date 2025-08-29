package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/rs/zerolog"
	"riverqueue.com/riverui"
)

type RiverHandler struct {
	handler *riverui.Handler
	logger  zerolog.Logger
}

func (h *RiverHandler) Handle(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())

	user := auth.GetUser(r.Context())

	if !user.IsAdmin() {
		errs.HTTPErrorResponse(w, h.logger, errs.E(errs.Unauthorized, fmt.Errorf("not an administrator")), requestID)

		return
	}

	h.handler.ServeHTTP(w, r)
}

func NewRiverHandler(handler *riverui.Handler, logger zerolog.Logger) *RiverHandler {
	mux := http.NewServeMux()
	mux.Handle("/river", handler)

	return &RiverHandler{
		handler: handler,
		logger:  logger,
	}
}
