package baas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/muhammadyunus/Restify-Service/internal/domain/entity"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// RESTGenerator generates HTTP handlers from endpoint DB bindings.
type RESTGenerator struct {
	introspector service.APIIntrospector
	logger       repository.Logger
}

// NewRESTGenerator creates a new REST API generator.
func NewRESTGenerator(introspector service.APIIntrospector, logger repository.Logger) service.RESTGenerator {
	return &RESTGenerator{
		introspector: introspector,
		logger:       logger,
	}
}

// GenerateHandler creates an HTTP handler for an endpoint.
func (g *RESTGenerator) GenerateHandler(ctx context.Context, ep *entity.Endpoint) (http.HandlerFunc, error) {
	switch ep.DBType {
	case entity.EndpointTypeTable:
		return g.generateTableHandler(ctx, ep), nil
	case entity.EndpointTypeFunction:
		return g.generateFunctionHandler(ctx, ep), nil
	case entity.EndpointTypeProcedure:
		return g.generateProcedureHandler(ctx, ep), nil
	default:
		return nil, fmt.Errorf("unsupported db_type: %s", ep.DBType)
	}
}

// generateTableHandler creates a handler for table-based endpoints.
func (g *RESTGenerator) generateTableHandler(ctx context.Context, ep *entity.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			g.handleSelectTable(w, r, ep)
		case http.MethodPost:
			g.handleInsertTable(w, r, ep)
		case http.MethodPut, http.MethodPatch:
			g.handleUpdateTable(w, r, ep)
		case http.MethodDelete:
			g.handleDeleteTable(w, r, ep)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleSelectTable handles GET requests for table data.
func (g *RESTGenerator) handleSelectTable(w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
	// TODO: Implement table select with query parameters
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message":    "table query handler",
		"table":      ep.TableName,
		"schema":     ep.Schema,
		"collection": ep.CollectionID,
	})
}

// handleInsertTable handles POST requests for table insertion.
func (g *RESTGenerator) handleInsertTable(w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
	// TODO: Implement table insert with JSON body parsing
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message":    "table insert handler",
		"table":      ep.TableName,
		"schema":     ep.Schema,
	})
}

// handleUpdateTable handles PUT/PATCH requests for table updates.
func (g *RESTGenerator) handleUpdateTable(w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
	// TODO: Implement table update
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		 "message":    "table update handler",
		"table":      ep.TableName,
		"schema":     ep.Schema,
	})
}

// handleDeleteTable handles DELETE requests for table records.
func (g *RESTGenerator) handleDeleteTable(w http.ResponseWriter, r *http.Request, ep *entity.Endpoint) {
	// TODO: Implement table delete
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message":    "table delete handler",
		"table":      ep.TableName,
		"schema":     ep.Schema,
	})
}

// generateFunctionHandler creates a handler for function-based endpoints.
func (g *RESTGenerator) generateFunctionHandler(ctx context.Context, ep *entity.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement function call handler
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"message":    "function call handler",
			"function":   ep.FuncName,
			"schema":     ep.Schema,
			"collection": ep.CollectionID,
		})
	}
}

// generateProcedureHandler creates a handler for procedure-based endpoints.
func (g *RESTGenerator) generateProcedureHandler(ctx context.Context, ep *entity.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement procedure call handler
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"message":    "procedure call handler",
			"procedure":  ep.FuncName,
			"schema":     ep.Schema,
			"collection": ep.CollectionID,
		})
	}
}

// ValidateBinding checks if the endpoint's DB binding is valid.
func (g *RESTGenerator) ValidateBinding(ctx context.Context, ep *entity.Endpoint) error {
	switch ep.DBType {
	case entity.EndpointTypeTable:
		return g.validateTableBinding(ctx, ep)
	case entity.EndpointTypeFunction, entity.EndpointTypeProcedure:
		return g.validateFunctionBinding(ctx, ep)
	default:
		return fmt.Errorf("unsupported db_type: %s", ep.DBType)
	}
}

// validateTableBinding checks if the table binding is valid.
func (g *RESTGenerator) validateTableBinding(ctx context.Context, ep *entity.Endpoint) error {
	if ep.TableName == "" {
		return fmt.Errorf("table_name is required for table endpoints")
	}
	if ep.Schema == "" {
		ep.Schema = "public"
	}

	// Verify table exists and get schema
	columns, err := g.introspector.GetTableSchema(ctx, ep.Schema, ep.TableName)
	if err != nil {
		return fmt.Errorf("validate table '%s.%s': %w", ep.Schema, ep.TableName, err)
	}

	if len(columns) == 0 {
		return fmt.Errorf("table '%s.%s' has no columns", ep.Schema, ep.TableName)
	}

	g.logger.Info(ctx, "validated table binding",
		"schema", ep.Schema,
		"table", ep.TableName,
		"columns", len(columns),
	)

	return nil
}

// validateFunctionBinding checks if the function/procedure binding is valid.
func (g *RESTGenerator) validateFunctionBinding(ctx context.Context, ep *entity.Endpoint) error {
	if ep.FuncName == "" {
		return fmt.Errorf("func_name is required for function/procedure endpoints")
	}
	if ep.Schema == "" {
		ep.Schema = "public"
	}

	// Verify function/procedure exists and get signature
	params, err := g.introspector.GetFunctionSignature(ctx, ep.Schema, ep.FuncName)
	if err != nil {
		return fmt.Errorf("validate %s '%s.%s': %w", ep.DBType, ep.Schema, ep.FuncName, err)
	}

	g.logger.Info(ctx, "validated function binding",
		"type", ep.DBType,
		"schema", ep.Schema,
		"name", ep.FuncName,
		"params", len(params),
	)

	return nil
}

// GeneratePath creates a RESTful path from an endpoint.
func (g *RESTGenerator) GeneratePath(ep *entity.Endpoint) string {
	// Use the endpoint's configured path if available
	if ep.Path != "" {
		return ep.Path
	}

	// Generate path from entity
	switch ep.DBType {
	case entity.EndpointTypeTable:
		return fmt.Sprintf("/%s/%s", ep.Schema, ep.TableName)
	case entity.EndpointTypeFunction, entity.EndpointTypeProcedure:
		return fmt.Sprintf("/%s/%s", ep.Schema, ep.FuncName)
	default:
		return "/unknown"
	}
}

// Compile-time checks
var _ service.RESTGenerator = (*RESTGenerator)(nil)
