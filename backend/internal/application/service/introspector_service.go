package service

import (
	"context"
	"fmt"

	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// IntrospectorService wraps the APIIntrospector for the application layer.
type IntrospectorService struct {
	introspector service.APIIntrospector
}

// NewIntrospectorService creates a new introspector service.
func NewIntrospectorService(introspector service.APIIntrospector) *IntrospectorService {
	return &IntrospectorService{introspector: introspector}
}

// DiscoverTables returns all user tables in a schema.
func (s *IntrospectorService) DiscoverTables(ctx context.Context, schema string) ([]service.TableInfo, error) {
	if schema == "" {
		schema = "public"
	}
	tables, err  := s.introspector.DiscoverTables(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("discover tables: %w", err)
	}
	return tables, nil
}

// DiscoverFunctions returns all user functions in a schema.
func (s *IntrospectorService) DiscoverFunctions(ctx context.Context, schema string) ([]service.FunctionInfo, error) {
	if schema == "" {
		schema = "public"
	}
	functions, err := s.introspector.DiscoverFunctions(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("discover functions: %w", err)
	}
	return functions, nil
}

// DiscoverProcedures returns all user procedures in a schema.
func (s *IntrospectorService) DiscoverProcedures(ctx context.Context, schema string) ([]service.ProcedureInfo, error) {
	if schema == "" {
		schema = "public"
	}
	procedures, err := s.introspector.DiscoverProcedures(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("discover procedures: %w", err)
	}
	return procedures, nil
}

// GetTableSchema returns column information for a table.
func (s *IntrospectorService) GetTableSchema(ctx context.Context, schema, table string) ([]service.ColumnSchema, error) {
	if schema == "" {
		schema = "public"
	}
	columns, err := s.introspector.GetTableSchema(ctx, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get table schema: %w", err)
	}
	return columns, nil
}

// GetFunctionSignature returns parameter info for a function/procedure.
func (s *IntrospectorService) GetFunctionSignature(ctx context.Context, schema, name string) ([]service.ParamSchema, error) {
	if schema == "" {
		schema = "public"
	}
	params, err := s.introspector.GetFunctionSignature(ctx, schema, name)
	if err != nil {
		return nil, fmt.Errorf("get function signature: %w", err)
	}
	return params, nil
}
