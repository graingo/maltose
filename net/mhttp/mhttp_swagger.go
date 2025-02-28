package mhttp

const (
	defaultSwaggerTemplate = `
<!DOCTYPE html>
<html>
    <head>
        <title>API Documentation</title>
        <meta charset="utf-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
        <style>
            body {
                margin: 0;
                background: #fafafa;
            }
            .swagger-ui .topbar {
                background: #2d3748;
                padding: 10px 0;
            }
            .swagger-ui .info {
                margin: 20px 0;
            }
            .swagger-ui .scheme-container {
                background: #fff;
                box-shadow: 0 1px 2px 0 rgba(0,0,0,0.1);
                position: sticky;
                top: 0;
                z-index: 100;
            }
            .swagger-ui .opblock {
                border-radius: 8px;
                box-shadow: 0 1px 3px 0 rgba(0,0,0,0.1);
                background: #fff;
                margin: 0 0 15px;
                border: none;
            }
            .swagger-ui .opblock .opblock-summary {
                padding: 10px;
            }
            .swagger-ui .opblock .opblock-summary-method {
                border-radius: 4px;
                min-width: 80px;
            }
            .swagger-ui .opblock-tag {
                font-size: 18px;
                font-weight: 600;
                margin: 20px 0 10px;
            }
            .swagger-ui .btn {
                box-shadow: 0 1px 3px 0 rgba(0,0,0,0.1);
            }
            .swagger-ui select {
                box-shadow: 0 1px 3px 0 rgba(0,0,0,0.1);
            }
            .swagger-ui .info .title {
                color: #2d3748;
            }
        </style>
    </head>
    <body>
        <div id="swagger-ui"></div>
        <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
        <script>
            window.onload = function() {
                window.ui = SwaggerUIBundle({
                    url: "%s",
                    dom_id: '#swagger-ui',
                    deepLinking: true,
                    presets: [
                        SwaggerUIBundle.presets.apis,
                        SwaggerUIBundle.SwaggerUIStandalonePreset
                    ],
                    layout: "BaseLayout",
                    docExpansion: "none",
                    defaultModelsExpandDepth: -1,
                    displayRequestDuration: true,
                    filter: true,
                    syntaxHighlight: {
                        activate: true,
                        theme: "agate"
                    }
                });
            }
        </script>
    </body>
</html>
`
)

// swaggerHandler Swagger UI 处理器
func (s *Server) swaggerHandler(r *Request) {
	template := defaultSwaggerTemplate
	if s.config.SwaggerTemplate != "" {
		template = s.config.SwaggerTemplate
	}
	r.Header("Content-Type", "text/html")
	if s.config.OpenapiPath == "" {
		r.String(200, "swagger path is empty")
		r.Abort()
	}
	r.String(200, template, s.config.OpenapiPath)
}
