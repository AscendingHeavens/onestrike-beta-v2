# OneStrike Framework (Go) V2.0.4

<p align="center">
  <img src="./onestrike-logo.jpg" alt="OneStrike Logo" width="200">
</p>


A minimal, high-performance Go web framework designed for clarity, speed, and decisive handling of HTTP requests. Inspired by Gin and Echo, OneStrike emphasizes explicit error handling, easy middleware composition, route groups, and ergonomic access to path and query parameters.

> [GO DOCS](https://pkg.go.dev/github.com/AscendingHeavens/onestrike/v2)  
Note: This repository was transferred to AscendingHeavens/onestrike. Due to Go module proxy caching, the official Go docs still point to the original path because the module was already published and tagged under the old repository.

---

⚠️ **This repository is no longer actively maintained.**  
It contains the final release series of **OneStrike v2.x.x**.  

Active development has moved to the new repository:  
👉 [github.com/AscendingHeavens/onestrike](https://github.com/AscendingHeavens/onestrike)

### Why the Change?
- New module path: `github.com/ascendingheavens/onestrike`
- Clean, lowercase import path
- Fresh semantic versioning history
- Future-proof for Go modules

### What Happens to This Repo?
- All previous tags (including `v2.x.x`) remain available on the Go proxy for reproducible builds.
- No new features, fixes, or releases will be added here.
- This repository is kept read-only for historical reference.

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
go get github.com/AscendingHeavens/onestrike/v2
```

---

## Usage

```go
package main

import (
	"net/http"

	onestrike "github.com/AscendingHeavens/onestrike/v2"
	"github.com/AscendingHeavens/onestrike/v2/middleware"
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

	app.GET("/html", func(c *onestrike.Context) *onestrike.Response {
		return c.HTML(200, "<h1>Serving HTML via OneStrike</h1>") // send file content as HTML
	})

	app.GET("/docs", func(c *onestrike.Context) *onestrike.Response {
		return c.Redirect(302, "https://google.com")
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
* More built-in middleware (JWT, etc.)
* Advanced profiling and metrics
* Testing
---
## Contributions

Contributions to OneStrike are welcome and encouraged! Here's how you can help:

### Reporting Issues

* Open an issue on GitHub if you find bugs, unexpected behavior, or have feature requests.
* Provide clear reproduction steps and examples whenever possible.

### Code Contributions

1. **Fork the repository**.
2. **Create a new branch** for your feature or bug fix: `git checkout -b feature/my-feature`.
3. **Write tests** for your changes.
4. **Make your changes** following the project's coding standards.
5. **Run all tests** locally to ensure nothing breaks.
6. **Commit your changes** with clear and descriptive messages.
7. **Push your branch** to your fork.
8. **Open a pull request** against the `main` branch of the main repository.

### Documentation

* Improve documentation, examples, or guides.
* Correct typos or clarify confusing explanations.

### Code Reviews

* Review open pull requests and provide feedback.
* Test new features and verify fixes.

### Style & Standards

* Follow Go idiomatic practices.
* Keep commits focused and concise.
* Write descriptive commit messages.

We appreciate all contributions, big or small, and they help make OneStrike better for everyone!



## License
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
