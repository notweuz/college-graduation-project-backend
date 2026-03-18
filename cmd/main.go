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
	//reviewDatabase := database.NewReviewDatabase(internal.Database)
	bookingDatabase := database.NewBookingDatabase(internal.Database)

	userService := service.NewUserService(userDatabase)
	authService := service.NewAuthService(userService)
	hallService := service.NewHallService(hallDatabase, bookingDatabase, userService)
	bookingService := service.NewBookingService(bookingDatabase, userService, hallService)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	hallHandler := handler.NewHallHandler(hallService)
	bookingHandler := handler.NewBookingHandler(bookingService)

	app := fiber.New(fiber.Config{
		ErrorHandler: handler.ErrorHandler,
	})
	app.Use(middleware.Logging())
	app.Use(cors.New())

	log.Info().Msg("Loading routers")
	r := router.NewRouter(app, authHandler, userHandler, hallHandler, bookingHandler)
	r.Setup()

	log.Info().Int("port", config.Cfg.AppPort).Msg("Starting server")
	err = app.Listen(fmt.Sprintf(":%d", config.Cfg.AppPort))
	if err != nil {
		log.Panic().Err(err).Msg("Error starting server")
	}
}
