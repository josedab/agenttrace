package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/agenttrace/agenttrace/api/docs"
)

// DocsHandler handles API documentation endpoints
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// RegisterRoutes registers documentation routes
func (h *DocsHandler) RegisterRoutes(app *fiber.App) {
	// Serve OpenAPI spec
	app.Get("/openapi.yaml", h.ServeOpenAPISpec)
	app.Get("/openapi.json", h.ServeOpenAPIJSON)

	// Serve Swagger UI
	app.Get("/docs", h.ServeSwaggerUI)
	app.Get("/docs/*", h.ServeSwaggerUI)

	// Serve ReDoc (alternative documentation)
	app.Get("/redoc", h.ServeReDoc)
}

// ServeOpenAPISpec serves the OpenAPI YAML specification
func (h *DocsHandler) ServeOpenAPISpec(c *fiber.Ctx) error {
	c.Set("Content-Type", "application/x-yaml")
	return c.Send(docs.OpenAPISpec)
}

// ServeOpenAPIJSON serves the OpenAPI spec as JSON (converted on the fly)
func (h *DocsHandler) ServeOpenAPIJSON(c *fiber.Ctx) error {
	// For simplicity, redirect to YAML - in production you'd convert
	return c.Redirect("/openapi.yaml")
}

// ServeSwaggerUI serves the Swagger UI HTML page
func (h *DocsHandler) ServeSwaggerUI(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AgentTrace API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
        .swagger-ui .topbar { display: none; }
        .swagger-ui .info { margin: 30px 0; }
        .swagger-ui .info .title { font-size: 36px; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                persistAuthorization: true,
                displayRequestDuration: true,
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
        };
    </script>
</body>
</html>`
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

// ServeReDoc serves the ReDoc documentation page
func (h *DocsHandler) ServeReDoc(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AgentTrace API Documentation - ReDoc</title>
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <redoc spec-url='/openapi.yaml'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}
