package handlers

import (
	"context"
	"net/http"

	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/service"
)

type UserHandler struct {
	service service.UserService
}

func (h *UserHandler) GetUserData(ctx context.Context, _ *http.Request, _ any) (*service.UserInfo, error) {
	const op errs.Op = "UserHandler.GetUserData"

	user := auth.GetUser(ctx)
	if user == nil {
		// FIXME: this might not be correct
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	return h.service.GetUserData(ctx, user)
}

func (h *UserHandler) GetTokens(ctx context.Context, _ *http.Request, _ any) ([]service.NadaToken, error) {
	const op errs.Op = "UserHandler.GetTokens"

	user := auth.GetUser(ctx)
	if user == nil {
		// FIXME: this might not be correct
		return nil, errs.E(errs.Unauthenticated, service.CodeNotLoggedIn, op, errs.Str("no user in context"))
	}

	return h.service.GetTokens(ctx, user)
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}
