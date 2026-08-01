package baas

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/goccy/go-yaml"
	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
)

// OpenAPIHandler handles GET /openapi.json and GET /openapi.yaml endpoints.
type OpenAPIHandler struct {
	generator      *APIDocGenerator
	endpointFetcher func(r *http.Request) ([]*entity.Endpoint, error)
}

// NewOpenAPIHandler creates a new OpenAPI handler.
func NewOpenAPIHandler(generator *APIDocGenerator, fetcher func(r *http.Request) ([]*entity.Endpoint, error)) *OpenAPIHandler {
	return &OpenAPIHandler{
		generator:       generator,
		endpointFetcher: fetcher,
	}
}

// GetJSONHandler returns the HTTP handler for serving OpenAPI JSON.
func (h *OpenAPIHandler) GetJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.serve(docFormatJSON, w, r)
	}
}

// GetYAMLHandler returns the HTTP handler for serving OpenAPI YAML.
func (h *OpenAPIHandler) GetYAMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.serve(docFormatYAML, w, r)
	}
}

func (h *OpenAPIHandler) serve(format docFormat, w http.ResponseWriter, r *http.Request) {
	endpoints, err := h.endpointFetcher(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch endpoints: %v", err), http.StatusInternalServerError)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Restify API"
	}
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "1.0.0"
	}
	baseURL := r.URL.Query().Get("baseurl")

	doc := h.generator.GenerateOpenAPI(r.Context(), endpoints, title, version, baseURL)

	switch format {
	case docFormatJSON:
		w.Header().Set("Content-Type", ContentTypeJSONValue)
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal JSON: %v", err), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(jsonBytes)
	case docFormatYAML:
		w.Header().Set("Content-Type", ContentTypeYAMLValue)
		yamlBytes, err := yaml.Marshal(doc)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal YAML: %v", err), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(yamlBytes)
	}
}
