package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/navikt/nada-backend/pkg/errs"

	"github.com/navikt/nada-backend/pkg/service"
)

type SlackHandler struct {
	service service.SlackService
}

func (h *SlackHandler) IsValidSlackChannel(_ context.Context, r *http.Request, _ any) (*service.IsValidSlackChannelResult, error) {
	const op errs.Op = "SlackHandler.IsValidSlackChannel"

	channelName := r.URL.Query().Get("channel")
	if channelName == "" {
		return nil, errs.E(errs.InvalidRequest, op, errs.Parameter("channel"), fmt.Errorf("channelName is required"))
	}

	err := h.service.IsValidSlackChannel(channelName)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.IsValidSlackChannelResult{
		IsValidSlackChannel: true,
	}, nil
}

func NewSlackHandler(service service.SlackService) *SlackHandler {
	return &SlackHandler{
		service: service,
	}
}
