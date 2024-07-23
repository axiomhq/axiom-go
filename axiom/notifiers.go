package axiom

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Notifier struct {
	// ID is the unique ID of the notifier.
	ID string `json:"id"`
	// Name is the name of the notifier.
	Name string `json:"name"`
	// DisabledUntil sets the notifier's disabled until property.
	DisabledUntil time.Time `json:"disabledUntil"`
	// Properties of the notifier.
	Properties NotifierProperties `json:"properties"`
	// CreatedAt is the time when the notifier was created.
	CreatedAt time.Time `json:"createdAt"`
	// CreatedBy is the user who created the notifier.
	CreatedBy string `json:"createdBy"`
}

type NotifierProperties struct {
	// Discord configuration.
	Discord *DiscordConfig `json:"discord,omitempty"`
	// DiscordWebhook configuration.
	DiscordWebhook *DiscordWebhookConfig `json:"discordWebhook,omitempty"`
	// Email configuration.
	Email *EmailConfig `json:"email,omitempty"`
	// OpsGenie configuration.
	Opsgenie *OpsGenieConfig `json:"opsgenie,omitempty"`
	// PagerDuty configuration.
	Pagerduty *PagerDutyConfig `json:"pagerduty,omitempty"`
	// Slack configuration.
	Slack *SlackConfig `json:"slack,omitempty"`
	// Webhook configuration.
	Webhook *WebhookConfig `json:"webhook,omitempty"`
	// MicrosoftTeams configuration.
	MicrosoftTeams *MicrosoftTeams `json:"microsoftTeams,omitempty"`
}

type DiscordConfig struct {
	// DiscordChannel is the channel to send the message to.
	DiscordChannel string `json:"discordChannel,omitempty"`
	// DiscordToken is the token to use for authentication.
	DiscordToken string `json:"discordToken,omitempty"`
}

type DiscordWebhookConfig struct {
	// DiscordWebhookURL is the URL to send the message to.
	DiscordWebhookURL string `json:"discordWebhookUrl,omitempty"`
}

type EmailConfig struct {
	// Emails to send the message to.
	Emails []string `json:"emails"`
}

type OpsGenieConfig struct {
	// APIKey is the API key to use for authentication.
	APIKey string `json:"apiKey,omitempty"`
	// IsEU indicates whether the OpsGenie instance is in the EU.
	IsEU bool `json:"isEU,omitempty"`
}

type PagerDutyConfig struct {
	// RoutingKey is the routing key to use for authentication.
	RoutingKey string `json:"routingKey,omitempty"`
	// Token is the token to use for authentication.
	Token string `json:"token,omitempty"`
}

type SlackConfig struct {
	// SlackChannel is the channel to send the message to.
	SlackURL string `json:"slackUrl,omitempty"`
}

type WebhookConfig struct {
	// URL is the URL to send the message to.
	URL string `json:"url,omitempty"`
}

type MicrosoftTeams struct {
	// URL is the URL to send the message to.
	URL string `json:"microsoftTeamsUrl,omitempty"`
}

// Axiom API Reference: /v2/notifiers
type NotifiersService service

// List all available notifiers.
func (s *NotifiersService) List(ctx context.Context) ([]*Notifier, error) {
	ctx, span := s.client.trace(ctx, "Notifiers.List")
	defer span.End()

	var res []*Notifier
	if err := s.client.Call(ctx, http.MethodGet, s.basePath, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get a notifier by id.
func (s *NotifiersService) Get(ctx context.Context, id string) (*Notifier, error) {
	ctx, span := s.client.trace(ctx, "Notifiers.Get", trace.WithAttributes(
		attribute.String("axiom.notifier_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Notifier
	if err := s.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Create a notifier with the given properties.
func (s *NotifiersService) Create(ctx context.Context, req Notifier) (*Notifier, error) {
	ctx, span := s.client.trace(ctx, "Notifiers.Create", trace.WithAttributes(
		attribute.String("axiom.param.name", req.Name),
	))
	defer span.End()

	var res Notifier
	if err := s.client.Call(ctx, http.MethodPost, s.basePath, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Update the notifier identified by the given id with the given properties.
func (s *NotifiersService) Update(ctx context.Context, id string, req Notifier) (*Notifier, error) {
	ctx, span := s.client.trace(ctx, "Notifiers.Update", trace.WithAttributes(
		attribute.String("axiom.notifier_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Notifier
	if err := s.client.Call(ctx, http.MethodPut, path, req, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the notifier identified by the given id.
func (s *NotifiersService) Delete(ctx context.Context, id string) error {
	ctx, span := s.client.trace(ctx, "Notifiers.Delete", trace.WithAttributes(
		attribute.String("axiom.notifier_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(s.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
