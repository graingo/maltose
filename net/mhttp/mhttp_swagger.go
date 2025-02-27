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
    </head>
    <body>
        <div id="swagger-ui"></div>
        <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
        <script>
            window.onload = function() {
                window.ui = SwaggerUIBundle({
                    url: "%s",
                    dom_id: '#swagger-ui',
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
