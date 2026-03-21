// @title College Graduation Project Backend API
// @version 1.0
// @description API для управления залами, бронированиями, отзывами и отчетами. Бекенд часть для диплома.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: Bearer <token>
//
//go:generate go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d .,../internal -o ../docs --parseDependency --parseInternal
package main

import (
	"college-graduation-project-backend/internal"
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/handler"
	"college-graduation-project-backend/internal/logger"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/router"
	"college-graduation-project-backend/internal/service"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.SetupLogger()
	err := config.SetupConfig()
	if err != nil {
		log.Panic().Err(err).Msg("Error loading config")
	}

	err = internal.SetupDatabase()
	if err != nil {
		log.Panic().Err(err).Msg("Error loading database")
	}

	userDatabase := database.NewUserDatabase(internal.Database)
	hallDatabase := database.NewHallDatabase(internal.Database)
	hallImageDatabase := database.NewHallImageDatabase(internal.Database)
	reviewDatabase := database.NewReviewDatabase(internal.Database)
	bookingDatabase := database.NewBookingDatabase(internal.Database)
	reportDatabase := database.NewReportDatabase(internal.Database)

	userService := service.NewUserService(userDatabase)
	authService := service.NewAuthService(userService)
	hallService := service.NewHallService(hallDatabase, bookingDatabase, userService, hallImageDatabase)
	bookingService := service.NewBookingService(bookingDatabase, userService, hallService)
	reviewService := service.NewReviewService(reviewDatabase, bookingService, userService)
	reportService := service.NewReportService(reportDatabase, userService)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	hallHandler := handler.NewHallHandler(hallService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	reportHandler := handler.NewReportHandler(reportService)

	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
	})
	app.Use(middleware.Logging())
	app.Use(cors.New())

	log.Info().Msg("Loading routers")
	r := router.NewRouter(app, authHandler, userHandler, hallHandler, bookingHandler, reviewHandler, reportHandler)
	r.Setup()

	log.Info().Int("port", config.Cfg.AppPort).Msg("Starting server")
	err = app.Listen(fmt.Sprintf(":%d", config.Cfg.AppPort))
	if err != nil {
		log.Panic().Err(err).Msg("Error starting server")
	}
}
