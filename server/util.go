package server

func (c *Context) writeResponse(code int, contentType string, body []byte) {
	if c.Handled {
		return
	}
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.WriteHeader(code)
	_, _ = c.Writer.Write(body)
	c.Handled = true
}
