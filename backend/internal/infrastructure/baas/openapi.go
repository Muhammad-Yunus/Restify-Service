package baas

// ParamSchema describes a parameter for a database function.
type ParamSchema struct {
	Name string
	Type string
	Mode string // IN, OUT, INOUT
}

// endpointSchema is an internal schema definition for endpoint params.
type endpointSchema struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Mode        string `json:"mode"`
	Description string `json:"description,omitempty"`
}

// OpenAPI represents a minimal OpenAPI v3 document.
type OpenAPI struct {
	OpenAPI  string               `json:"openapi" yaml:"openapi"`
	Info     Info                 `json:"info" yaml:"info"`
	Servers  []Server             `json:"servers,omitempty" yaml:"servers,omitempty"`
	Tags     []Tag                `json:"tags,omitempty" yaml:"tags,omitempty"`
	Paths    map[string]*PathItem `json:"paths" yaml:"paths"`
	Security []SecurityReq        `json:"security,omitempty" yaml:"security,omitempty"`
}

// Info holds API metadata.
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// Server describes an API server.
type Server struct {
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Tag groups API endpoints.
type Tag struct {
	Name string `json:"name"`
	Desc string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Paths is the paths map: path string -> PathItem.
type Paths map[string]*PathItem

// PathKey is a typed string for path keys to prevent mixing with other strings.
type PathKey string

// PathItem represents all HTTP methods for a single path.
type PathItem struct {
	Get        *Operation        `json:"get,omitempty" yaml:"get,omitempty"`
	Post       *Operation        `json:"post,omitempty" yaml:"post,omitempty"`
	Put        *Operation        `json:"put,omitempty" yaml:"put,omitempty"`
	Delete     *Operation        `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch      *Operation        `json:"patch,omitempty" yaml:"patch,omitempty"`
	Parameters []CommonParameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation describes a single HTTP operation.
type Operation struct {
	Tags        []string             `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string               `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string               `json:"operationId" yaml:"operationId"`
	Parameters  []Param              `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Request     *RequestBody         `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]*Response `json:"responses" yaml:"responses"`
	Security    []SecurityReq        `json:"security,omitempty" yaml:"security,omitempty"`
}

// CommonParameter is a parameter that can appear in PathItem.Parameters or Operation.Parameters.
type CommonParameter struct {
	Name        string  `json:"name" yaml:"name"`
	In          string  `json:"in" yaml:"in"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// Param is an operation-level parameter.
type Param struct {
	Name        string  `json:"name" yaml:"name"`
	In          string  `json:"in" yaml:"in"` // path, query, header, cookie
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// RequestBody describes the expected request body.
type RequestBody struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]*MediaType `json:"content" yaml:"content"`
}

// MediaType describes content type details.
type MediaType struct {
	Schema *Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// Response describes a single response.
type Response struct {
	Description string                      `json:"description" yaml:"description"`
	Content     map[string]*MediaType       `json:"content,omitempty" yaml:"content,omitempty"`
	Headers     map[string]*CommonParameter `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// Schema describes the shape of a parameter or response value.
type Schema struct {
	Type        string             `json:"type,omitempty" yaml:"type,omitempty"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Items       *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required    []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Ref         string             `json:"$$ref,omitempty" yaml:"$$ref,omitempty"`
	AnyOf       []*Schema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
}

// SecurityReq defines a security requirement.
type SecurityReq struct {
	BearerToken []string `json:"bearerAuth,omitempty" yaml:"bearerAuth,omitempty"`
}

// Component defines reusable schemas.
type Component struct {
	Schemas map[string]*Schema `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

// securityRequirement is the key used for bearer auth in security requirements.
const securityKeyBearerAuth = "bearerAuth"

// httpMethodToOpenAPI maps HTTP method to OpenAPI operation field.
func methodToOpenAPIOp(method string) string {
	switch method {
	case "GET":
		return "get"
	case "POST":
		return "post"
	case "PUT":
		return "put"
	case "DELETE":
		return "delete"
	case "PATCH":
		return "patch"
	case "HEAD":
		return "head"
	case "OPTIONS":
		return "options"
	default:
		return "unknown"
	}
}

// openAPISchemaType maps a DB type string to an OpenAPI JSON Schema type.
func openAPISchemaType(dbType string) string {
	switch dbType {
	case "text", "varchar", "char", "uuid", "serializable", "tsvector":
		return "string"
	case "integer", "smallint", "bigint", "serial", "bigserial", "int2", "int4", "int8":
		return "integer"
	case "boolean", "bool":
		return "boolean"
	case "json", "jsonb", "xml":
		return "object"
	case "numeric", "decimal", "real", "double precision", "float4", "float8":
		return "number"
	default:
		return "string"
	}
}

// buildParamSchema creates a Schema from endpoint parameters.
func buildParamSchema(params []ParamSchema) *Schema {
	if len(params) == 0 {
		return nil
	}
	props := make(map[string]*Schema, len(params))
	var required []string
	for _, p := range params {
		props[p.Name] = &Schema{
			Type:        openAPISchemaType(p.Type),
			Description: "Input parameter",
		}
		if p.Mode == "IN" || p.Mode == "INOUT" {
			required = append(required, p.Name)
		}
	}
	return &Schema{
		Type:       "object",
		Properties: props,
		Required:   required,
	}
}
