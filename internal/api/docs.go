package api

import (
	"fmt"
	"net/http"
)

// swagger serves the Swagger UI page, which reads /openapi.yaml.
func (h *handler) swagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
	<html>
	<head>
	<title>CacheLab API</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css">
	</head>
	<body>
	<div id="swagger-ui"></div>
	<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
	<script>
		SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger-ui" });
	</script>
	</body>
	</html>`)
}
