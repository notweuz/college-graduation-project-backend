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
	hallHandler *handler.HallHandler
}

func NewRouter(app *fiber.App, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, hallHandler *handler.HallHandler) *Router {
	return &Router{
		app:         app,
		authHandler: authHandler,
		userHandler: userHandler,
		hallHandler: hallHandler,
	}
}

func (r *Router) Setup() {
	api := r.app.Group("/api")
	admin := api.Group("/admin")

	r.setupAuthRoutes(api)
	r.setupUserRoutes(api)
	r.setupHallRoutes(api)
	r.setupAdminHallRoutes(admin)
}

func (r *Router) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Post("/register", r.authHandler.Register)
	auth.Post("/login", r.authHandler.Login)
}

func (r *Router) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users", middleware.Protected())

	users.Get("/me", r.userHandler.GetProfile)
	users.Patch("/me", r.userHandler.UpdateProfile)
}

func (r *Router) setupHallRoutes(api fiber.Router) {
	halls := api.Group("/halls")

	halls.Get("/", r.hallHandler.GetAllHalls)
	halls.Get("/:id", r.hallHandler.GetHallById)
}

func (r *Router) setupAdminHallRoutes(api fiber.Router) {
	halls := api.Group("/halls", middleware.Protected())

	halls.Post("/", r.hallHandler.Create)
}
