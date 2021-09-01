package models

type QPWebhookV2 struct {
	ChannelPost QPMessageV2 `json:"channel_post"`
}
