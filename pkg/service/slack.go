package service

type IsValidSlackChannelResult struct {
	IsValidSlackChannel bool `json:"isValidSlackChannel"`
}

type SlackAPI interface {
	SendSlackNotification(channel, message string) error
	IsValidSlackChannel(name string) error
}

type SlackService interface {
	IsValidSlackChannel(name string) error
}
