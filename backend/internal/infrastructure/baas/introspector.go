package baas

import (
	"context"
	"fmt"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// PostgresIntrospector inspects PostgreSQL schemas and discovers tables, functions, and procedures.
type PostgresIntrospector struct {
	db repository.DB
}

// NewPostgreSQLIntrospector creates a new PostgreSQL schema introspector.
func NewPostgreSQLIntrospector(db repository.DB) service.APIIntrospector {
	return &PostgresIntrospector{db: db}
}

// DiscoverTables returns all user tables in a schema.
func (i *PostgresIntrospector) DiscoverTables(ctx context.Context, schema string) ([]service.TableInfo, error) {
	query := `
		SELECT table_schema, table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		  AND (table_name NOT LIKE 'pg_%' OR table_schema != 'pg_catalog')
		ORDER BY table_name
	`
	rows, err := i.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("discover tables: %w", err)
	}

	var tables []service.TableInfo
	for _, row := range rows {
		schemaName, _ := row["table_schema"].(string)
		tableName, _ := row["table_name"].(string)
		tables = append(tables, service.TableInfo{
			Schema: schemaName,
			Name:   tableName,
		})
	}

	return tables, nil
}

// DiscoverFunctions returns all user functions in a schema.
func (i *PostgresIntrospector) DiscoverFunctions(ctx context.Context, schema string) ([]service.FunctionInfo, error) {
	query := `
		SELECT 
			n.nspname AS schema_name,
			p.proname AS function_name,
			p.prorettype::regtype::text AS return_type
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1
		  AND has_function_privilege(p.oid, 'execute')
		  AND p.prokind = 'f'
		ORDER BY p.proname
	`
	rows, err := i.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("discover functions: %w", err)
	}

	var functions []service.FunctionInfo
	for _, row := range rows {
		schemaName, _ := row["schema_name"].(string)
		funcName, _ := row["function_name"].(string)
		returnType, _ := row["return_type"].(string)
		functions = append(functions, service.FunctionInfo{
			Schema:     schemaName,
			Name:       funcName,
			ReturnType: returnType,
		})
	}

	return functions, nil
}

// DiscoverProcedures returns all user procedures in a schema.
func (i *PostgresIntrospector) DiscoverProcedures(ctx context.Context, schema string) ([]service.ProcedureInfo, error) {
	query := `
		SELECT 
			n.nspname AS schema_name,
			p.proname AS procedure_name
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname = $1
		  AND has_function_privilege(p.oid, 'execute')
		  AND p.prokind = 'p'
		ORDER BY p.proname
	`
	rows, err := i.db.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("discover procedures: %w", err)
	}

	var procedures []service.ProcedureInfo
	for _, row := range rows {
		schemaName, _ := row["schema_name"].(string)
		procName, _ := row["procedure_name"].(string)
		procedures = append(procedures, service.ProcedureInfo{
			Schema: schemaName,
			Name:   procName,
		})
	}

	return procedures, nil
}

// GetTableSchema returns column information for a table.
func (i *PostgresIntrospector) GetTableSchema(ctx context.Context, schema, table string) ([]service.ColumnSchema, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_primary,
			c.column_default::text AS default_value
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.table_schema, ku.table_name, ku.column_name
			FROM information_schema.key_column_usage ku
			JOIN information_schema.table_constraints tc 
				ON tc.constraint_schema = ku.table_schema 
				AND tc.constraint_name = ku.constraint_name
				AND tc.constraint_type = 'PRIMARY KEY'
		) pk ON pk.table_schema = c.table_schema 
			AND pk.table_name = c.table_name 
			AND pk.column_name = c.column_name
		WHERE c.table_schema = $1
		  AND c.table_name = $2
		ORDER BY c.ordinal_position
	`
	rows, err := i.db.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("get table schema: %w", err)
	}

	var columns []service.ColumnSchema
	for _, row := range rows {
		colName, _ := row["column_name"].(string)
		colType, _ := row["data_type"].(string)
		isNullable, _ := row["is_nullable"].(string)
		isPrimary, _ := row["is_primary"].(bool)
		defaultVal := "null"
		if v, ok := row["default_value"]; ok && v != nil {
			defaultVal, _ = v.(string)
		}

		columns = append(columns, service.ColumnSchema{
			Name:       colName,
			Type:       colType,
			IsNullable: isNullable == "YES",
			IsPrimary:  isPrimary,
			Default:    &defaultVal,
		})
	}

	return columns, nil
}

// GetFunctionSignature returns parameter information for a function/procedure.
func (i *PostgresIntrospector) GetFunctionSignature(ctx context.Context, schema, name string) ([]service.ParamSchema, error) {
	query := `
		SELECT 
			p.arg_number,
			p.arg_name,
			p.arg_type::regtype::text AS arg_type,
			p.arg_mode
		FROM (
			SELECT 
				unnest(p.proargnames) AS arg_name,
				unnest(p.proargmodes) AS arg_mode,
				unnest(p.proallargtypes) AS arg_type,
				unnest(p.proargnumbers) AS arg_number
			FROM pg_proc p
			JOIN pg_namespace n ON n.oid = p.pronamespace
			WHERE n.nspname = $1
			  AND p.proname = $2
			  AND has_function_privilege(p.oid, 'execute')
		) p
		ORDER BY p.arg_number
	`
	rows, err := i.db.Query(ctx, query, schema, name)
	if err != nil {
		return nil, fmt.Errorf("get function signature: %w", err)
	}

	var params []service.ParamSchema
	for _, row := range rows {
		argName, _ := row["arg_name"].(string)
		argType, _ := row["arg_type"].(string)
		argMode, _ := row["arg_mode"].(string)
		if argMode == "" {
			argMode = "IN"
		}
		params = append(params, service.ParamSchema{
			Name: argName,
			Type: argType,
			Mode: argMode,
		})
	}

	return params, nil
}

// Compile-time check for service.APIIntrospector interface
var _ service.APIIntrospector = (*PostgresIntrospector)(nil)
