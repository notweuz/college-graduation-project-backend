package internal

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/handler"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/router"
	"college-graduation-project-backend/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type App struct {
	Fiber *fiber.App
}

func NewApp(db *gorm.DB) *App {
	userDatabase := database.NewUserDatabase(db)
	hallDatabase := database.NewHallDatabase(db)
	imageDatabase := database.NewImageDatabase(db)
	reviewDatabase := database.NewReviewDatabase(db)
	bookingDatabase := database.NewBookingDatabase(db)
	reportDatabase := database.NewReportDatabase(db)
	userAgreementDatabase := database.NewUserAgreementDatabase(db)

	userService := service.NewUserService(userDatabase)
	imageService := service.NewImageService(imageDatabase, userDatabase, hallDatabase)
	authService := service.NewAuthService(userService)
	hallService := service.NewHallService(hallDatabase, bookingDatabase, userService)
	bookingService := service.NewBookingService(bookingDatabase, userService, hallService)
	reviewService := service.NewReviewService(reviewDatabase, bookingService, userService)
	reportService := service.NewReportService(reportDatabase, userService)
	userAgreementService := service.NewUserAgreementService(userAgreementDatabase, userService)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, imageService)
	hallHandler := handler.NewHallHandler(hallService, imageService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	reportHandler := handler.NewReportHandler(reportService)
	userAgreementHandler := handler.NewUserAgreementHandler(userAgreementService)

	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
	})
	app.Use(middleware.Logging())
	app.Use(cors.New())

	log.Info().Msg("Loading routers")
	r := router.NewRouter(app, authHandler, userHandler, hallHandler, bookingHandler, reviewHandler, reportHandler, userAgreementHandler)
	r.Setup()

	return &App{Fiber: app}
}
