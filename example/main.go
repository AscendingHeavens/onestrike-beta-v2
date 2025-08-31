package main

import (
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

	// Route group
	v1 := app.Group("/api/v1")
	v1.GET("/users/:id", func(c *onestrike.Context) *onestrike.Response {
		id := c.Params["id"]
		return &onestrike.Response{Success: true, Message: "User found", Details: map[string]string{"id": id}, Code: 200}
	})

	auth := app.Group("auth")
	auth.POST("/signup", Signup)

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
	return &onestrike.Response{
		Message: "Signup",
	}
}
