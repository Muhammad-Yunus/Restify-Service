package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/muhammadyunus/Restify-Service/internal/application/service"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/presentation/http/dto"
)

// IntrospectHandler handles API introspection requests.
type IntrospectHandler struct {
	introspectService *service.IntrospectorService
}

// NewIntrospectHandler creates a new introspect handler.
func NewIntrospectHandler(introspectService *service.IntrospectorService) *IntrospectHandler {
	return &IntrospectHandler{introspectService: introspectService}
}

// @Summary		Discover tables
// @Description	Discover all tables in a PostgreSQL schema
// @Tags			introspection
// @Produce		json
// @Param			schema	path		string	true	"Schema name"	default(public)
// @Success		200		{object}	map[string]interface{}
// @Failure		500		{object}	map[string]interface{}
// @Router			/api/v1/introspect/schemas/{schema}/tables [get]
// @Security		BearerAuth
func (h *IntrospectHandler) DiscoverTables(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	tables, err := h.introspectService.DiscoverTables(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tables})
}

// @Summary		Discover functions
// @Description	Discover all functions in a PostgreSQL schema
// @Tags			introspection
// @Produce		json
// @Param			schema	path		string	true	"Schema name"	default(public)
// @Success		200		{object}	map[string]interface{}
// @Failure		500		{object}	map[string]interface{}
// @Router			/api/v1/introspect/schemas/{schema}/functions [get]
// @Security		BearerAuth
func (h *IntrospectHandler) DiscoverFunctions(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	functions, err := h.introspectService.DiscoverFunctions(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": functions})
}

// @Summary		Discover procedures
// @Description	Discover all procedures in a PostgreSQL schema
// @Tags			introspection
// @Produce		json
// @Param			schema	path		string	true	"Schema name"	default(public)
// @Success		200		{object}	map[string]interface{}
// @Failure		500		{object}	map[string]interface{}
// @Router			/api/v1/introspect/schemas/{schema}/procedures [get]
// @Security		BearerAuth
func (h *IntrospectHandler) DiscoverProcedures(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	procedures, err := h.introspectService.DiscoverProcedures(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": procedures})
}

// @Summary		Get table schema
// @Description	Get column information for a table
// @Tags			introspection
// @Produce		json
// @Param			schema	path		string	true	"Schema name"	default(public)
// @Param			table	path		string	true	"Table name"
// @Success		200		{object}	map[string]interface{}
// @Failure		500		{object}	map[string]interface{}
// @Router			/api/v1/introspect/schemas/{schema}/tables/{table} [get]
// @Security		BearerAuth
func (h *IntrospectHandler) GetTableSchema(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")
	table := c.Param("table")

	if table == "" {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Missing table name", "table parameter is required"))
		return
	}

	columns, err := h.introspectService.GetTableSchema(c.Request.Context(), schema, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"schema":  schema,
		"table":   table,
		"columns": columns,
	}})
}

// @Summary		Get function signature
// @Description	Get parameter information for a function/procedure
// @Tags			introspection
// @Produce		json
// @Param			schema	path		string	true	"Schema name"	default(public)
// @Param			name	path		string	true	"Function/Procedure name"
// @Success		200		{object}	map[string]interface{}
// @Failure		500		{object}	map[string]interface{}
// @Router			/api/v1/introspect/schemas/{schema}/functions/{name} [get]
// @Security		BearerAuth
func (h *IntrospectHandler) GetFunctionSignature(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, dto.ProblemDetail(http.StatusBadRequest, "Missing function name", "name parameter is required"))
		return
	}

	params, err := h.introspectService.GetFunctionSignature(c.Request.Context(), schema, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"schema": schema,
		"name":   name,
		"params": params,
	}})
}
