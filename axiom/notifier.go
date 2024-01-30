package axiom

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/url"
)

type Notifier struct {
	DisabledUntil string              `json:"disabledUntil,omitempty"`
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Properties    *NotifierProperties `json:"properties"`
	Type          string              `json:"type"`
}

type NotifierProperties struct {
	Discord        *DiscordConfig        `json:"discord,omitempty"`
	DiscordWebhook *DiscordWebhookConfig `json:"discordWebhook,omitempty"`
	Email          *EmailConfig          `json:"email,omitempty"`
	Opsgenie       *OpsGenieConfig       `json:"opsgenie,omitempty"`
	Pagerduty      *PagerDutyConfig      `json:"pagerduty,omitempty"`
	Slack          *SlackConfig          `json:"slack,omitempty"`
	Webhook        *WebhookConfig        `json:"webhook,omitempty"`
}

type DiscordConfig struct {
	DiscordChannel string `json:"DiscordChannel,omitempty"`
	DiscordToken   string `json:"DiscordToken,omitempty"`
}

type DiscordWebhookConfig struct {
	DiscordWebhookURL string `json:"DiscordWebhookUrl,omitempty"`
}

type EmailConfig struct {
	Emails []string `json:"Emails"`
}

type OpsGenieConfig struct {
	APIKey string `json:"ApiKey,omitempty"`
	IsEU   bool   `json:"IsEU,omitempty"`
}

type PagerDutyConfig struct {
	RoutingKey string `json:"RoutingKey,omitempty"`
	Token      string `json:"Token,omitempty"`
}

type SlackConfig struct {
	SlackURL string `json:"SlackUrl,omitempty"`
}

type WebhookConfig struct {
	URL string `json:"Url,omitempty"`
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
		attribute.String("axiom.Notifier_id", id),
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

	path, err := url.JoinPath(s.basePath, "/", id)
	if err != nil {
		return spanError(span, err)
	}

	if err := s.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
