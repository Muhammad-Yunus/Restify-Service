package baas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// BodyMapper maps HTTP request body fields to database parameters.
type BodyMapper struct {
	mapping map[string]BodyMappingRule
}

// BodyMappingRule defines a single source→target mapping with optional transform.
type BodyMappingRule struct {
	SourceField string `json:"source_field"`
	TargetParam string `json:"target_param"`
	Transform   string `json:"transform,omitempty"`
}

// NewBodyMapper parses optional body-mapping rules from the endpoint's config blob.
func NewBodyMapper(mappingJSON []byte) (*BodyMapper, error) {
	bm := &BodyMapper{}
	if len(mappingJSON) == 0 {
		return bm, nil
	}

	var rules []BodyMappingRule
	if err := json.Unmarshal(mappingJSON, &rules); err != nil {
		return nil, fmt.Errorf("parse body mapping: %w", err)
	}

	bm.mapping = make(map[string]BodyMappingRule, len(rules))
	for _, rule := range rules {
		bm.mapping[rule.SourceField] = rule
	}

	return bm, nil
}

// MapBody reads the request body as JSON and applies custom mapping rules.
func (bm *BodyMapper) MapBody(r *http.Request) (map[string]any, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var raw map[string]any
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON body: %w", err)
	}

	return bm.applyMapping(raw), nil
}

func (bm *BodyMapper) applyMapping(raw map[string]any) map[string]any {
	if bm == nil || len(bm.mapping) == 0 {
		return raw
	}

	result := make(map[string]any, len(raw))
	for sourceField, value := range raw {
		rule, hasRule := bm.mapping[sourceField]
		if hasRule {
			result[rule.TargetParam] = applyTransform(value, rule.Transform)
		} else {
			result[sourceField] = value
		}
	}
	return result
}

func applyTransform(value any, transform string) any {
	switch transform {
	case "lowercase":
		if s, ok := value.(string); ok {
			return strings.ToLower(s)
		}
	case "uppercase":
		if s, ok := value.(string); ok {
			return strings.ToUpper(s)
		}
	case "trim":
		if s, ok := value.(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return value
}