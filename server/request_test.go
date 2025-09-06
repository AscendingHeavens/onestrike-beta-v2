package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestContextWithBody(method, contentType, body string) *Context {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	return &Context{
		Request: req,
		Writer:  w,
		Params:  make(map[string]string),
	}
}

func TestBind_JSON_Pass(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":"Rishi"}`)
	var p payload

	err := c.Bind(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestBind_JSON_Fail(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":`) // invalid
	var p payload

	err := c.Bind(&p)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Writer.(*httptest.ResponseRecorder).Result().StatusCode)
}

func TestBind_XML_Pass(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</Name></payload>`)
	var p payload

	err := c.Bind(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestBind_XML_Fail(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</payload>`) // malformed
	var p payload

	err := c.Bind(&p)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Writer.(*httptest.ResponseRecorder).Result().StatusCode)
}

func TestBind_UnsupportedContentType(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "text/plain", `hello`)
	var p payload

	err := c.Bind(&p)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, c.Writer.(*httptest.ResponseRecorder).Result().StatusCode)
}

func TestBindJSON_ValidJSON_Passes(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":"Rishi"}`)
	var p payload

	err := c.BindJSON(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestBindJSON_InvalidJSON_Returns400(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":`) // broken JSON
	var p payload

	err := c.BindJSON(&p)
	assert.Error(t, err)

	// Verify 400 response
	rec := c.Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
	body := rec.Body.String()
	assert.Contains(t, body, "Invalid JSON body")
}

func TestBindXML_ValidXML_Passes(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</Name></payload>`)
	var p payload

	err := c.BindXML(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestBindXML_InvalidXML_Returns400(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</payload>`) // malformed
	var p payload

	err := c.BindXML(&p)
	assert.Error(t, err)

	rec := c.Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
	body := rec.Body.String()
	assert.Contains(t, body, "Invalid XML body")
}

func TestShouldBind_JSON_CallsJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":"Rishi"}`)
	var p payload

	err := c.ShouldBind(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBind_XML_CallsXML(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</Name></payload>`)
	var p payload

	err := c.ShouldBind(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBind_Form_CallsForm(t *testing.T) {
	type payload struct {
		Name string `form:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/x-www-form-urlencoded", `Name=Rishi`)
	var p payload

	err := c.ShouldBind(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBind_Multipart_CallsMultipart(t *testing.T) {
	type payload struct {
		Name string `form:"Name"`
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("Name", "Rishi")
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	c := newTestContextWithBody(http.MethodPost, writer.FormDataContentType(), body.String())
	var p payload

	err := c.ShouldBind(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBind_UnsupportedContentType_ReturnsError(t *testing.T) {
	type payload struct {
		Name string
	}
	c := newTestContextWithBody(http.MethodPost, "text/plain", "hello")
	var p payload

	err := c.ShouldBind(&p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported Content-Type")
}

func TestShouldBind_MissingContentType_ReturnsError(t *testing.T) {
	type payload struct {
		Name string
	}
	c := newTestContextWithBody(http.MethodPost, "", `{"name":"Rishi"}`)
	var p payload

	err := c.ShouldBind(&p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing Content-Type header")
}

func TestShouldBindJSON_ValidJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":"Rishi"}`)
	var p payload

	err := c.ShouldBindJSON(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBindJSON_InvalidJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/json", `{"name":`)
	var p payload

	err := c.ShouldBindJSON(&p)
	assert.Error(t, err)
}

func TestShouldBindXML_ValidXML(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</Name></payload>`)
	var p payload

	err := c.ShouldBindXML(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBindXML_InvalidXML(t *testing.T) {
	type payload struct {
		Name string `xml:"Name"`
	}
	c := newTestContextWithBody(http.MethodPost, "application/xml", `<payload><Name>Rishi</payload>`) // malformed
	var p payload

	err := c.ShouldBindXML(&p)
	assert.Error(t, err)
}

func TestShouldBindForm_ValidForm(t *testing.T) {
	type payload struct {
		Name string `form:"Name"`
		Age  string `form:"Age"`
	}

	form := "Name=Rishi&Age=21"
	c := newTestContextWithBody(http.MethodPost, "application/x-www-form-urlencoded", form)
	var p payload

	err := c.ShouldBindForm(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
	assert.Equal(t, "21", p.Age)
}

func TestShouldBindForm_EmptyForm(t *testing.T) {
	type payload struct {
		Name string `form:"Name"`
	}

	// malformed form won't cause error; PostForm will be empty
	c := newTestContextWithBody(http.MethodPost, "application/x-www-form-urlencoded", "\x00invalid")
	var p payload

	err := c.ShouldBindForm(&p)
	assert.NoError(t, err)
	assert.Equal(t, "", p.Name)
}

func TestShouldBindMultipart_ValidMultipart(t *testing.T) {
	type payload struct {
		Name string `form:"Name"`
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("Name", "Rishi")
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	c := newTestContextWithBody(http.MethodPost, writer.FormDataContentType(), body.String())
	var p payload

	err := c.ShouldBindMultipart(&p)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", p.Name)
}

func TestShouldBindMultipart_ParseError(t *testing.T) {
	type payload struct {
		Name string `form:"Name"`
	}

	c := newTestContextWithBody(http.MethodPost, "multipart/form-data; boundary=invalid", "bad body")
	var p payload

	err := c.ShouldBindMultipart(&p)
	assert.Error(t, err)
}

func TestParam_ReturnsValueOrEmpty(t *testing.T) {
	c := &Context{Params: map[string]string{"id": "123"}}
	assert.Equal(t, "123", c.Param("id"))
	assert.Equal(t, "", c.Param("missing"))

	// nil Params
	c.Params = nil
	assert.Equal(t, "", c.Param("id"))
}

func TestQuery_ReturnsFirstValueOrEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/search?q=golang&lang=go", nil)
	c := &Context{Request: req}

	assert.Equal(t, "golang", c.Query("q"))
	assert.Equal(t, "go", c.Query("lang"))
	assert.Equal(t, "", c.Query("missing"))

	// nil Request
	c.Request = nil
	assert.Equal(t, "", c.Query("q"))
}

func TestQueryArray_ReturnsAllValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/filter?tag=go&tag=web", nil)
	c := &Context{Request: req}

	values := c.QueryArray("tag")
	assert.Equal(t, []string{"go", "web"}, values)
	assert.Nil(t, c.QueryArray("missing"))

	// nil Request
	c.Request = nil
	assert.Nil(t, c.QueryArray("tag"))
}

func TestFormFile_ReturnsFileOrError(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = fw.Write([]byte("hello"))
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close FormFile writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	c := &Context{Request: req}
	file, fh, err := c.FormFile("file")
	assert.NoError(t, err)
	assert.Equal(t, "test.txt", fh.Filename)
	content, _ := io.ReadAll(file)
	assert.Equal(t, []byte("hello"), content)

	_, _, err = c.FormFile("missing")
	assert.Error(t, err)
}

func TestFormValue_ReturnsValueOrEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("username=Rishi"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := &Context{Request: req}

	assert.Equal(t, "Rishi", c.FormValue("username"))
	assert.Equal(t, "", c.FormValue("missing"))
}

func TestBindForm_Success(t *testing.T) {
	body := strings.NewReader("username=Rishi&age=21")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := &Context{Request: req}
	dest := make(map[string]string)

	err := c.BindForm(dest)
	assert.NoError(t, err)
	assert.Equal(t, "Rishi", dest["username"])
	assert.Equal(t, "21", dest["age"])
}

func TestBindForm_MissingContentType(t *testing.T) {
	body := strings.NewReader("username=Rishi")
	req := httptest.NewRequest(http.MethodPost, "/", body)

	c := &Context{Request: req}
	dest := make(map[string]string)

	err := c.BindForm(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing content type header")
}

func TestBindForm_InvalidContentType(t *testing.T) {
	body := strings.NewReader("username=Rishi")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")

	c := &Context{Request: req}
	dest := make(map[string]string)

	err := c.BindForm(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid content type")
}

func TestBindForm_EmptyBody(t *testing.T) {
	body := strings.NewReader("")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := &Context{Request: req}
	dest := make(map[string]string)

	err := c.BindForm(dest)
	assert.NoError(t, err)
	assert.Empty(t, dest)
}

func TestBindFormAll_Success(t *testing.T) {
	body := strings.NewReader("tag=go&tag=web&user=Rishi")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := &Context{Request: req}

	values, err := c.BindFormAll()
	assert.NoError(t, err)
	assert.Equal(t, []string{"go", "web"}, values["tag"])
	assert.Equal(t, []string{"Rishi"}, values["user"])
}

func TestBindFormAll_MissingContentType(t *testing.T) {
	body := strings.NewReader("user=Rishi")
	req := httptest.NewRequest(http.MethodPost, "/", body)

	c := &Context{Request: req}
	values, err := c.BindFormAll()
	assert.Nil(t, values)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing content type header")
}

func TestBindFormAll_InvalidContentType(t *testing.T) {
	body := strings.NewReader("user=Rishi")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/json")

	c := &Context{Request: req}
	values, err := c.BindFormAll()
	assert.Nil(t, values)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid content type")
}

func TestBindFormAll_EmptyBody(t *testing.T) {
	body := strings.NewReader("")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := &Context{Request: req}
	values, err := c.BindFormAll()
	assert.NoError(t, err)
	assert.Empty(t, values)
}
