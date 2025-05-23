package slack

import (
	"fmt"
	"strings"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	slackapi "github.com/slack-go/slack"
)

type slackAPI struct {
	webhookURL      string
	token           string
	channelOverride string
	api             *slackapi.Client
}

var _ service.SlackAPI = &slackAPI{}

func (a *slackAPI) SendSlackNotification(channel, message string) error {
	const op = "slackAPI.SendSlackNotification"

	// Global override, usually used in dev
	if len(a.channelOverride) > 0 {
		channel = a.channelOverride
	}

	_, _, _, err := a.api.SendMessage(channel, slackapi.MsgOptionText(message, false))
	if err != nil {
		return errs.E(errs.IO, service.CodeSlack, op, err)
	}

	return nil
}

func (a *slackAPI) IsValidSlackChannel(name string) error {
	const op = "slackAPI.IsValidSlackChannel"

	c := ""
	for i := 0; i < 10; i++ {
		chn, nc, e := a.api.GetConversations(&slackapi.GetConversationsParameters{
			Cursor:          c,
			ExcludeArchived: true,
			Types:           []string{"public_channel"},
			Limit:           1000,
		})
		if e != nil {
			return errs.E(errs.IO, service.CodeSlack, op, e)
		}

		for _, cn := range chn {
			if strings.EqualFold(cn.Name, name) {
				return nil
			}
		}

		if nc == "" {
			return errs.E(errs.NotExist, service.CodeSlack, op, fmt.Errorf("channel %s not found", name), service.ParamChannel)
		}

		c = nc
	}

	return errs.E(errs.Internal, service.CodeSlack, op, fmt.Errorf("too many channels to search"))
}

// NewSlackAPI creates a Slack client.
//
// If channelOverride is set, all notifications will be sent to that channel
// instead of their specified destination.
func NewSlackAPI(webhookURL, token, channelOverride string) *slackAPI {
	return &slackAPI{
		webhookURL:      webhookURL,
		token:           token,
		channelOverride: channelOverride,
		api:             slackapi.New(token),
	}
}
