package router

import (
	"college-graduation-project-backend/internal/handler"
	"college-graduation-project-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

type Router struct {
	app         *fiber.App
	authHandler *handler.AuthHandler
	userHandler *handler.UserHandler
}

func NewRouter(app *fiber.App, authHandler *handler.AuthHandler, userHandler *handler.UserHandler) *Router {
	return &Router{
		app:         app,
		authHandler: authHandler,
		userHandler: userHandler,
	}
}

func (r *Router) Setup() {
	api := r.app.Group("/api")

	r.setupAuthRoutes(api)
	r.setupUserRoutes(api)
}

func (r *Router) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Post("/register", r.authHandler.Register)
	auth.Post("/login", r.authHandler.Login)
}

func (r *Router) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users", middleware.Protected())

	users.Get("/me", r.userHandler.GetProfile)
	//users.Patch("/me", r.userHandler.UpdateProfile)
}
