package baas

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func introspectToParamSchema(schemas []service.ParamSchema) []ParamSchema {
	result := make([]ParamSchema, len(schemas))
	for i, s := range schemas {
		result[i] = ParamSchema{
			Name: s.Name,
			Type: s.Type,
			Mode: s.Mode,
		}
	}
	return result
}

// APIDocGenerator generates OpenAPI v3 documents from endpoint entities.
type APIDocGenerator struct {
	introspector service.APIIntrospector
	logger       repository.Logger
}

// NewAPIDocGenerator creates a new API documentation generator.
func NewAPIDocGenerator(introspector service.APIIntrospector, logger repository.Logger) *APIDocGenerator {
	return &APIDocGenerator{
		introspector: introspector,
		logger:       logger,
	}
}

// GenerateOpenAPI produces a full OpenAPI document for the given endpoints.
func (g *APIDocGenerator) GenerateOpenAPI(ctx context.Context, endpoints []*entity.Endpoint,
	title, version, baseURL string) *OpenAPI {

	doc := &OpenAPI{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       title,
			Version:     version,
			Description: fmt.Sprintf("Auto-generated API documentation for %s v%s", title, version),
		},
		Servers: []Server{
			{URL: baseURL, Description: "API Server"},
		},
		Paths:    make(map[string]*PathItem),
		Tags:     make([]Tag, 0),
		Security: []SecurityReq{{BearerToken: []string{}}},
	}

	// Group by collection name for tags
	collections := make(map[string]string) // endpointID -> collectionName
	for _, ep := range endpoints {
		colls := collections[ep.ID.String()]
		if colls == "" && ep.Collection != nil {
			collections[ep.ID.String()] = ep.Collection.Name
		}
	}
	seenTags := make(map[string]string) // name -> desc
	for id, name := range collections {
		if name != "" && seenTags[name] == "" {
			seenTags[name] = id
			doc.Tags = append(doc.Tags, Tag{
				Name: name,
				Desc: name,
			})
		}
	}

	for _, ep := range endpoints {
		g.addEndpointToDoc(ctx, doc, ep)
	}

	return doc
}

func (g *APIDocGenerator) addEndpointToDoc(ctx context.Context, doc *OpenAPI, ep *entity.Endpoint) {
	path := sanitizePath(ep.Path)
	op := g.buildOperation(ctx, ep)
	if op == nil {
		return
	}

	pathItem := doc.Paths[path]
	if pathItem == nil {
		pathItem = &PathItem{}
	}
	switch ep.Method {
	case "GET":
		pathItem.Get = op
	case "POST":
		pathItem.Post = op
	case "PUT":
		pathItem.Put = op
	case "DELETE":
		pathItem.Delete = op
	case "PATCH":
		pathItem.Patch = op
	}
	doc.Paths[path] = pathItem
}

func (g *APIDocGenerator) buildOperation(ctx context.Context, ep *entity.Endpoint) *Operation {
	method := ep.Method
	if method == "" {
		method = "GET"
	}

	op := &Operation{
		OperationID: method + "_" + sanitizePath(ep.Path),
		Summary:     ep.Name,
		Parameters:  buildParams(ep),
		Responses: map[string]*Response{
			"200": {
				Description: "Successful response",
				Content: map[string]*MediaType{
					"application/json": {Schema: &Schema{Type: "object"}},
				},
			},
		},
	}

	if ep.Description != nil && *ep.Description != "" {
		op.Description = *ep.Description
	}

	if ep.Collection != nil {
		op.Tags = []string{ep.Collection.Name}
	}

	policy := ep.GetSecurityPolicy()
	if policy.AuthRequired {
		op.Security = []SecurityReq{{BearerToken: []string{}}}
	}

	if ep.DBType == entity.EndpointTypeFunction || ep.DBType == entity.EndpointTypeProcedure {
		schema := g.resolveInputSchema(ctx, ep)
		if schema != nil {
			op.Request = &RequestBody{
				Content: map[string]*MediaType{
					"application/json": {Schema: schema},
				},
			}
		}
		op.Responses["200"].Content["application/json"].Schema = buildOutputSchema(ep)
	}

	return op
}

func (g *APIDocGenerator) resolveInputSchema(ctx context.Context, ep *entity.Endpoint) *Schema {
	if g.introspector == nil {
		return nil
	}
	funcName := ep.FuncName
	if funcName == "" {
		return nil
	}
	schema, err := g.introspector.GetFunctionSignature(ctx, ep.Schema, funcName)
	if err != nil {
		g.logger.Error(ctx, "introspect function signature", "error", err)
		return nil
	}
	return buildParamSchema(
		introspectToParamSchema(schema),
	)
}

// sanitizePath ensures the path starts with a leading slash and is URL-safe.
func sanitizePath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	return path
}

// buildParams creates parameter definitions from endpoint config.
func buildParams(ep *entity.Endpoint) []Param {
	var params []Param

	if ep.AuthHeader != "" && ep.AuthHeader != "Authorization" {
		params = append(params, Param{
			Name:        "Authorization",
			In:          "header",
			Description: "Authentication header",
			Required:    true,
			Schema:      &Schema{Type: "string"},
		})
	}

	if len(ep.ParamHeaders) > 0 {
		var mapping map[string]string
		if err := jsonUnmarshal(ep.ParamHeaders, &mapping); err == nil {
			for header, paramName := range mapping {
				params = append(params, Param{
					Name:        paramName,
					In:          "header",
					Description: fmt.Sprintf("Mapped from header %s", header),
					Schema:      &Schema{Type: "string"},
				})
			}
		}
	}

	if len(ep.BodyMappingJSON) > 0 {
		var rules []BodyMappingRule
		if err := jsonUnmarshal(ep.BodyMappingJSON, &rules); err == nil {
			for _, rule := range rules {
				params = append(params, Param{
					Name:        rule.SourceField,
					In:          "body",
					Description: fmt.Sprintf("Maps to %s", rule.TargetParam),
				})
			}
		}
	}

	return params
}

func buildOutputSchema(ep *entity.Endpoint) *Schema {
	if ep.DBType == entity.EndpointTypeTable {
		return &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"data":   {Type: "object"},
				"status": {Type: "string"},
			},
		}
	}
	return &Schema{Type: "object"}
}
