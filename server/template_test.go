package server_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ascending-Heavens/onestrike/server"
	"github.com/stretchr/testify/assert"
)

func createTempTemplate(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.html")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	assert.NoError(t, err)
	return tmpFile
}

// 1. Normal render success
func TestContext_Render_Success(t *testing.T) {
	templatePath := createTempTemplate(t, "Hello {{.Name}}")
	tr := server.NewTemplateRenderer(templatePath, false, nil)

	rec := httptest.NewRecorder()
	ctx := &server.Context{
		Writer:    rec,
		Templates: tr,
	}

	resp := ctx.Render(ctx.Templates, http.StatusOK, "test.html", map[string]string{"Name": "Rishi"})
	assert.True(t, resp.Success)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, rec.Body.String(), "Hello Rishi")
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
}

// 2. Already handled response
func TestContext_Render_AlreadyHandled(t *testing.T) {
	templatePath := createTempTemplate(t, "Hello {{.Name}}")
	tr := server.NewTemplateRenderer(templatePath, false, nil)

	rec := httptest.NewRecorder()
	ctx := &server.Context{
		Writer:    rec,
		Handled:   true,
		Templates: tr,
	}

	resp := ctx.Render(ctx.Templates, http.StatusOK, "test.html", nil)
	assert.False(t, resp.Success)
	assert.Equal(t, "Response already handled", resp.Message)
	assert.Equal(t, http.StatusOK, resp.Code)
}

// 3. Template render error (non-existent template)
func TestContext_Render_TemplateError(t *testing.T) {
	templatePath := createTempTemplate(t, "Hello {{.Name}}")
	tr := server.NewTemplateRenderer(templatePath, false, nil)

	rec := httptest.NewRecorder()
	ctx := &server.Context{
		Writer:    rec,
		Templates: tr,
	}

	resp := ctx.Render(ctx.Templates, http.StatusOK, "doesnotexist.html", nil)
	assert.False(t, resp.Success)
	assert.Equal(t, 500, resp.Code)
	assert.Contains(t, rec.Body.String(), "Template error")
}

// 4. Dev mode reload (reloads template on every render)
func TestContext_Render_DevModeReload(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.html")

	// initial content
	err := os.WriteFile(tmpFile, []byte("Hello {{.Name}}"), 0644)
	assert.NoError(t, err)

	tr := server.NewTemplateRenderer(tmpFile, true, nil) // devMode = true

	rec1 := httptest.NewRecorder()
	ctx1 := &server.Context{
		Writer:    rec1,
		Templates: tr,
	}

	resp1 := ctx1.Render(ctx1.Templates, http.StatusOK, "test.html", map[string]string{"Name": "Rishi"})
	assert.True(t, resp1.Success)
	assert.Contains(t, rec1.Body.String(), "Hello Rishi")

	// modify template
	err = os.WriteFile(tmpFile, []byte("Bye {{.Name}}"), 0644)
	assert.NoError(t, err)

	rec2 := httptest.NewRecorder()
	ctx2 := &server.Context{
		Writer:    rec2,
		Templates: tr,
	}

	resp2 := ctx2.Render(ctx2.Templates, http.StatusOK, "test.html", map[string]string{"Name": "Rishi"})
	assert.True(t, resp2.Success)
	assert.Contains(t, rec2.Body.String(), "Bye Rishi") // devMode reload worked
}

// 1. Render success for TemplateRenderer
func TestTemplateRenderer_Render_Success(t *testing.T) {
	templatePath := createTempTemplate(t, "Hello {{.Name}}")
	tr := server.NewTemplateRenderer(templatePath, false, nil)

	rec := httptest.NewRecorder()
	err := tr.Render(rec, "test.html", map[string]string{"Name": "Rishi"})
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "Hello Rishi")
}

// 2. Render error when template does not exist
func TestTemplateRenderer_Render_TemplateNotFound(t *testing.T) {
	templatePath := createTempTemplate(t, "Hello {{.Name}}")
	tr := server.NewTemplateRenderer(templatePath, false, nil)

	rec := httptest.NewRecorder()
	err := tr.Render(rec, "nonexistent.html", nil)
	assert.Error(t, err)
}

// 3. mustLoad panics on invalid template
func TestTemplateRenderer_MustLoad_InvalidTemplate(t *testing.T) {
	// create a file with invalid template syntax
	tmpFile := createTempTemplate(t, "{{.Name") // missing closing brace
	assert.Panics(t, func() {
		server.NewTemplateRenderer(tmpFile, false, nil)
	})
}

// 4. Dev mode reload actually reloads changed templates
func TestTemplateRenderer_DevModeReload(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.html")

	// initial content
	err := os.WriteFile(tmpFile, []byte("Hello {{.Name}}"), 0644)
	assert.NoError(t, err)

	tr := server.NewTemplateRenderer(tmpFile, true, nil) // devMode = true

	// First render
	rec1 := httptest.NewRecorder()
	err = tr.Render(rec1, "test.html", map[string]string{"Name": "Rishi"})
	assert.NoError(t, err)
	assert.Contains(t, rec1.Body.String(), "Hello Rishi")

	// change template content
	err = os.WriteFile(tmpFile, []byte("Bye {{.Name}}"), 0644)
	assert.NoError(t, err)

	// Second render should reload new template
	rec2 := httptest.NewRecorder()
	err = tr.Render(rec2, "test.html", map[string]string{"Name": "Rishi"})
	assert.NoError(t, err)
	assert.Contains(t, rec2.Body.String(), "Bye Rishi")
}
