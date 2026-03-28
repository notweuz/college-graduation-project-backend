package router

import (
	"college-graduation-project-backend/internal/handler"
	"college-graduation-project-backend/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

type Router struct {
	app                  *fiber.App
	authHandler          *handler.AuthHandler
	userHandler          *handler.UserHandler
	hallHandler          *handler.HallHandler
	bookingHandler       *handler.BookingHandler
	reviewHandler        *handler.ReviewHandler
	reportHandler        *handler.ReportHandler
	userAgreementHandler *handler.UserAgreementHandler
}

func NewRouter(app *fiber.App, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, hallHandler *handler.HallHandler, bookingHandler *handler.BookingHandler, reviewHandler *handler.ReviewHandler, reportHandler *handler.ReportHandler, userAgreementHandler *handler.UserAgreementHandler) *Router {
	return &Router{
		app:                  app,
		authHandler:          authHandler,
		userHandler:          userHandler,
		hallHandler:          hallHandler,
		bookingHandler:       bookingHandler,
		reviewHandler:        reviewHandler,
		reportHandler:        reportHandler,
		userAgreementHandler: userAgreementHandler,
	}
}

func (r *Router) Setup() {
	api := r.app.Group("/api")
	admin := api.Group("/admin")

	r.setupAuthRoutes(api)
	r.setupPublicRoutes(api)
	r.setupUserRoutes(api)
	r.setupHallRoutes(api)
	r.setupImageRoutes(api)
	r.setupBookingRoutes(api)
	r.setupReviewRoutes(api)
	r.setupAdminHallRoutes(admin)
	r.setupAdminBookingRoutes(admin)
	r.setupAdminReviewRoutes(admin)
	r.setupAdminReportRoutes(admin)
	r.setupAdminUserAgreementRoutes(admin)
}

func (r *Router) setupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Post("/register", r.authHandler.Register)
	auth.Post("/login", r.authHandler.Login)
}

func (r *Router) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users", middleware.Protected())

	users.Get("/me", r.userHandler.GetProfile)
	users.Get("/me/role", r.userHandler.GetRole)
	users.Get("/me/avatar", r.userHandler.GetAvatar)
	users.Get("/:id", r.userHandler.GetPublicProfile)
	users.Patch("/me", r.userHandler.UpdateProfile)
	users.Put("/me/avatar", r.userHandler.UploadAvatar)
	users.Delete("/me/avatar", r.userHandler.DeleteAvatar)
}

func (r *Router) setupPublicRoutes(api fiber.Router) {
	public := api.Group("/public")
	public.Get("/user-agreement", r.userAgreementHandler.GetPublic)
}

func (r *Router) setupHallRoutes(api fiber.Router) {
	halls := api.Group("/halls")

	halls.Get("/", r.hallHandler.GetAllHalls)
	halls.Get("/:id", r.hallHandler.GetHallById)
	halls.Get("/:id/availability", r.hallHandler.GetHallAvailability)
	halls.Get("/:id/reviews", r.reviewHandler.GetByHallID)
}

func (r *Router) setupImageRoutes(api fiber.Router) {
	images := api.Group("/images")

	images.Get("/:filename", r.hallHandler.ServeImage)
}

func (r *Router) setupBookingRoutes(api fiber.Router) {
	bookings := api.Group("/bookings", middleware.Protected())

	bookings.Post("/", r.bookingHandler.Create)
	bookings.Get("/my", r.bookingHandler.GetAllFromUser)
	bookings.Get("/:id", r.bookingHandler.GetByID)
	bookings.Get("/:hall_id/calculate-price", r.bookingHandler.CalculatePrice)
	bookings.Post("/:id/review", r.reviewHandler.Create)
	bookings.Delete("/:id", r.bookingHandler.DeleteByAuthor)
}

func (r *Router) setupAdminHallRoutes(api fiber.Router) {
	halls := api.Group("/halls", middleware.Protected())

	halls.Post("/", r.hallHandler.Create)
	halls.Patch("/:id", r.hallHandler.Update)
	halls.Delete("/:id", r.hallHandler.Delete)
	halls.Post("/:id/images", r.hallHandler.UploadImage)
}

func (r *Router) setupAdminBookingRoutes(api fiber.Router) {
	bookings := api.Group("/bookings", middleware.Protected())

	bookings.Get("/", r.bookingHandler.GetAll)
	bookings.Patch("/:id", r.bookingHandler.Update)
	bookings.Get("/:id", r.bookingHandler.GetByID)
}

func (r *Router) setupReviewRoutes(api fiber.Router) {
	reviews := api.Group("/reviews", middleware.Protected())

	reviews.Patch("/:id", r.reviewHandler.Update)
}

func (r *Router) setupAdminReviewRoutes(api fiber.Router) {
	reviews := api.Group("/reviews", middleware.Protected())

	reviews.Get("/", r.reviewHandler.GetAllAdmin)
	reviews.Delete("/:id", r.reviewHandler.DeleteAdmin)
}

func (r *Router) setupAdminReportRoutes(api fiber.Router) {
	reports := api.Group("/reports", middleware.Protected())

	reports.Get("/sales", r.reportHandler.GetSalesReport)
	reports.Get("/sales/pdf", r.reportHandler.GetSalesReportPDF)
	reports.Get("/halls-load", r.reportHandler.GetHallsLoadReport)
	reports.Get("/halls-load/pdf", r.reportHandler.GetHallsLoadReportPDF)
	reports.Get("/clients", r.reportHandler.GetClientsReport)
	reports.Get("/clients/pdf", r.reportHandler.GetClientsReportPDF)
	reports.Get("/bookings-dynamics", r.reportHandler.GetBookingsDynamicsReport)
	reports.Get("/bookings-dynamics/pdf", r.reportHandler.GetBookingsDynamicsReportPDF)
}

func (r *Router) setupAdminUserAgreementRoutes(api fiber.Router) {
	api.Put("/user-agreement", middleware.Protected(), r.userAgreementHandler.UpdateAdmin)
}
