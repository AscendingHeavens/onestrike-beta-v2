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
