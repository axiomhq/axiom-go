package axiom

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Annotation represents an Annotation returned from the Axiom API.
type Annotation struct {
	ID          string    `json:"id"`
	Datasets    []string  `json:"datasets"`
	Time        time.Time `json:"time"`
	EndTime     time.Time `json:"end_time"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Type        string    `json:"type"`
}

// AnnotationCreateRequest is the request payload for creating a new Annotation.
type AnnotationCreateRequest struct {
	// Required:
	Datasets []string `json:"datasets"`
	Type     string   `json:"type"`
	// Optional:
	Time        time.Time `json:"time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
}

// AnnotationUpdateRequest is the request payload for updating an Annotation.
type AnnotationUpdateRequest struct {
	Datasets    []string  `json:"datasets,omitempty"`
	Time        time.Time `json:"time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Type        string    `json:"type,omitempty"`
}

// AnnotationsService handles communication with the dataset related operations
// of the Axiom API.
//
// Axiom API Reference: /v2/annotations
type AnnotationsService service

// Create a new annotation.
func (a *AnnotationsService) Create(ctx context.Context, annotation *AnnotationCreateRequest) (*Annotation, error) {
	ctx, span := a.client.trace(ctx, "Annotations.Create")
	defer span.End()

	var res Annotation
	if err := a.client.Call(ctx, http.MethodPost, a.basePath, annotation, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// AnnotationsFilter filters annotations on the [`List`] method.
type AnnotationsFilter struct {
	Datasets []string   `url:"datasets"`
	Start    *time.Time `url:"start"`
	End      *time.Time `url:"start"`
}

// List annotations.
func (a *AnnotationsService) List(ctx context.Context, filter *AnnotationsFilter) ([]*Annotation, error) {
	ctx, span := a.client.trace(ctx, "Annotations.List", trace.WithAttributes())
	defer span.End()

	path, err := AddURLOptions(a.basePath, filter)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res []*Annotation
	if err := a.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return res, nil
}

// Get the Annotation with the given ID.
func (a *AnnotationsService) Get(ctx context.Context, id string) (*Annotation, error) {
	ctx, span := a.client.trace(ctx, "Annotations.Get", trace.WithAttributes(
		attribute.String("axiom.annotation_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(a.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Annotation
	if err := a.client.Call(ctx, http.MethodGet, path, nil, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

func (a *AnnotationsService) Update(ctx context.Context, id string, annotation *AnnotationUpdateRequest) (*Annotation, error) {
	ctx, span := a.client.trace(ctx, "Annotations.Update")
	defer span.End()

	path, err := url.JoinPath(a.basePath, id)
	if err != nil {
		return nil, spanError(span, err)
	}

	var res Annotation
	if err := a.client.Call(ctx, http.MethodPut, path, annotation, &res); err != nil {
		return nil, spanError(span, err)
	}

	return &res, nil
}

// Delete the Annotation with the given ID.
func (a *AnnotationsService) Delete(ctx context.Context, id string) error {
	ctx, span := a.client.trace(ctx, "Annotations.Delete", trace.WithAttributes(
		attribute.String("axiom.annotation_id", id),
	))
	defer span.End()

	path, err := url.JoinPath(a.basePath, id)
	if err != nil {
		return spanError(span, err)
	}

	if err := a.client.Call(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return spanError(span, err)
	}

	return nil
}
