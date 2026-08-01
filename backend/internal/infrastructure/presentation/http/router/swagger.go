package router

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

// swaggerTemplateFunc returns a Gin handler that serves a simple Swagger UI
// that points to the generated swagger.json.
func swaggerTemplateFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ForgeBase API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/swagger/doc.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            })
        }
    </script>
</body>
</html>`

		var buf bytes.Buffer
		if err := template.Must(template.New("swagger-ui").Parse(tmpl)).Execute(&buf, nil); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	}
}

// swaggerDocHandler serves the swagger.json content.
func swaggerDocHandler(ext string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ext == "json" {
			c.File("./docs/swagger.json")
			return
		}

		c.File("./docs/swagger.yaml")
	}
}

// RegisterSwaggerRoutes registers Swagger UI and API documentation routes.
// It serves both the swagger.json and swagger.yaml at /swagger/doc.json
// and /swagger/doc.yaml respectively, and the UI at /swagger/.
func RegisterSwaggerRoutes(r *gin.Engine) {
	if r == nil {
		return
	}

	// Serve swagger JSON and YAML docs as static routes (must be registered
	// before the wildcard handler).
	r.GET("/swagger/doc.json", swaggerDocHandler("json"))
	r.GET("/swagger/doc.yaml", swaggerDocHandler("yaml"))

	// Serve the swagger UI (handles /swagger, /swagger/, /swagger/index.html).
	r.GET("/swagger", swaggerTemplateFunc())
	r.GET("/swagger/", swaggerTemplateFunc())
	r.GET("/swagger/index.html", swaggerTemplateFunc())

	// Redirect legacy swagger-json path
	r.GET("/swagger-json", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/doc.json")
	})
}
