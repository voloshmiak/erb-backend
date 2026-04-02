package controller

import "net/http"

type DocsController struct {
}

func NewDocsController() *DocsController {
	return &DocsController{}
}

// ServeHTTP godoc
// @Summary     API Documentation
// @Description Serves the interactive API reference UI
// @Tags        System
// @Produce     html
// @Success     200  {string}  string  "HTML page"
// @Router      /docs [get]
func (h *DocsController) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `
    <!DOCTYPE html>
    <html>
      <head>
        <title>ERB API Reference</title>
        <meta charset="utf-8" />
        <style>body { margin: 0; }</style>
      </head>
      <body>
        <script id="api-reference" data-url="/swagger/doc.json"></script>
        <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
      </body>
    </html>
    `
	_, err := w.Write([]byte(html))
	if err != nil {
		http.Error(w, "Failed to load API reference", http.StatusInternalServerError)
		return
	}
}
