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

// DiscoverTables returns all tables in a schema.
func (h *IntrospectHandler) DiscoverTables(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	tables, err := h.introspectService.DiscoverTables(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tables})
}

// DiscoverFunctions returns all functions in a schema.
func (h *IntrospectHandler) DiscoverFunctions(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	functions, err := h.introspectService.DiscoverFunctions(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": functions})
}

// DiscoverProcedures returns all procedures in a schema.
func (h *IntrospectHandler) DiscoverProcedures(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	procedures, err := h.introspectService.DiscoverProcedures(c.Request.Context(), schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ProblemDetail(http.StatusInternalServerError, "Internal error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": procedures})
}

// GetTableSchema returns column information for a table.
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

// GetFunctionSignature returns parameter information for a function/procedure.
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
