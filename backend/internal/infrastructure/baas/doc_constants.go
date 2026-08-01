package baas

// Constants for documentation generation.
const (
	ContentTypeJSON       = "Content-Type"
	ContentTypeYAML       = "Content-Type"
	ContentTypeJSONValue  = "application/json"
	ContentTypeYAMLValue  = "application/yaml"
)

type docFormat string

const (
	docFormatJSON docFormat = "json"
	docFormatYAML docFormat = "yaml"
)
