# OneStrike Framework (Go)

A minimal, high-performance Go web framework designed for clarity, speed, and decisive handling of HTTP requests. Inspired by Gin and Echo, OneStrike emphasizes explicit error handling, easy middleware composition, route groups, and ergonomic access to path and query parameters.

---

## Features

* Routing with HTTP methods: GET, POST, PUT, PATCH, DELETE
* Route groups with middleware inheritance
* Global and conditional middleware (use on specific routes or patterns)
* Explicit error handling via `*Response` objects
* Automatic JSON response encoding
* Panic recovery middleware
* Profiling middleware with memory stats and execution time
* Path parameters (`/users/:id`) via `c.Param("id")`
* Query parameters via `c.Query("key")`
* Body binding with fail-fast: `Bind` / `BindJSON`
* Optional error-return binding: `ShouldBind` / `ShouldBindJSON`

---

## Installation

```bash
go get github.com/Rishi-Mishra0704/OneStrike
```

---

## Usage

```go
package main

import (
	"net/http"

	onestrike "github.com/Rishi-Mishra0704/OneStrike"
	"github.com/Rishi-Mishra0704/OneStrike/middleware"
)

func main() {
	app := onestrike.New()

	// Global middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.ProfilingMiddleware())

	// Conditional middleware
	app.UseIf("/api/v1/*", AuthMiddleware())

	// Top-level route
	app.GET("/ping", func(c *onestrike.Context) *onestrike.Response {
		return &onestrike.Response{Success: true, Message: "pong", Code: 200}
	})
	app.GET("/search", func(c *onestrike.Context) *onestrike.Response {
		q := c.Query("q")
		return &onestrike.Response{
			Success: true,
			Message: "Query received",
			Details: map[string]string{"query": q},
			Code:    200,
		}
	})

	// Route group
	v1 := app.Group("/api/v1")
	v1.GET("/users/:id", func(c *onestrike.Context) *onestrike.Response {
		id := c.Param("id")
		return &onestrike.Response{Success: true, Message: "User found", Details: map[string]string{"id": id}, Code: 200}
	})

	auth := app.Group("/auth")
	auth.POST("/signup", Signup)
	auth.GET("/users/:id", func(c *onestrike.Context) *onestrike.Response {
		id := c.Param("id")
		return &onestrike.Response{
			Success: true,
			Message: "User found",
			Details: map[string]string{"id": id},
			Code:    200,
		}
	})

	// Start server
	app.Start(":8080")
}

func AuthMiddleware() onestrike.Middleware {
	return func(next onestrike.HandlerFunc) onestrike.HandlerFunc {
		return func(c *onestrike.Context) *onestrike.Response {
			if c.Request.Header.Get("Authorization") == "" {
				return &onestrike.Response{
					Success: false,
					Message: "Unauthorized",
					Code:    401,
				}
			}
			return next(c)
		}
	}
}

func Signup(ctx *onestrike.Context) *onestrike.Response {
	var req struct{}
	if err := ctx.BindJSON(&req); err != nil {
		return nil
	}
	return &onestrike.Response{
		Message: "Signup successful",
		Success: true,
		Code:    http.StatusOK,
	}
}

```

---

## Middleware Examples

### Logger

Logs request method, path, response status, and execution time.

### Recovery

Catches panics and prevents server crash.

### Profiling

Logs request duration and memory usage.

---

## Response & Error Handling

Handlers return `*onestrike.Response` objects:

```go
return &onestrike.Response{
    Success: true,
    Message: "OK",
    Details: map[string]string{"id": id},
    Code: 200,
}
```

Fail-fast body binding automatically sends `400 Bad Request`:

```go
ctx.BindJSON(&req) // on failure, JSON 400 response sent
```

For custom error handling, use `ShouldBindJSON` and return your own response.

---

## Query & Path Params

```go
q := c.Query("q")       // /search?q=golang
id := c.Param("id")      // /users/:id
```

---

## Future Enhancements

* Form/url-encoded body support
* Validation helpers
* More built-in middleware (CORS, JWT, etc.)
* Advanced profiling and metrics

---

## License

MIT
