package server

import (
	"html/template"
	"net/http"
	"sync"
)

type TemplateRenderer struct {
	templates *template.Template
	funcs     template.FuncMap
	mu        sync.RWMutex
	pattern   string
	devMode   bool
}

// NewTemplateRenderer initializes a renderer that parses templates from the given glob pattern.
// Example: NewTemplateRenderer("views/*.html", true) for dev mode.
func NewTemplateRenderer(pattern string, devMode bool, funcs template.FuncMap) *TemplateRenderer {
	tr := &TemplateRenderer{
		funcs:   funcs,
		pattern: pattern,
		devMode: devMode,
	}
	tr.mustLoad()
	return tr
}

// mustLoad loads and parses all templates, panicking if any error occurs (fail fast).
func (tr *TemplateRenderer) mustLoad() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	parsed, err := template.New("").Funcs(tr.funcs).ParseGlob(tr.pattern)
	if err != nil {
		panic("failed to parse templates: " + err.Error())
	}
	tr.templates = parsed
}

// Render renders a template with the provided name and data into the response writer.
func (tr *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}) error {
	if tr.devMode {
		// Reload templates every time in dev mode
		tr.mustLoad()
	}

	tr.mu.RLock()
	defer tr.mu.RUnlock()

	return tr.templates.ExecuteTemplate(w, name, data)
}

// Render helper on Context so users can do c.Render("index.html", data)
func (c *Context) Render(renderer *TemplateRenderer, code int, name string, data interface{}) *Response {
	if c.Handled {
		return &Response{Success: false, Message: "Response already handled", Code: code}
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(code)

	err := renderer.Render(c.Writer, name, data)
	if err != nil {
		// Fall back to error
		http.Error(c.Writer, "Template error: "+err.Error(), http.StatusInternalServerError)
		return &Response{Success: false, Message: "Template render error", Code: 500}
	}

	c.Handled = true
	return &Response{Success: true, Message: "Template rendered: " + name, Code: code}
}
