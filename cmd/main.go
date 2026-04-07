// @title College Graduation Project Backend API
// @version 1.1
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
	"college-graduation-project-backend/internal/logger"
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	logger.SetupLogger()
	err := config.SetupConfig()
	if err != nil {
		log.Panic().Err(err).Msg("Error loading config")
	}

	db, err := internal.SetupDatabase()
	if err != nil {
		log.Panic().Err(err).Msg("Error loading database")
	}

	app := internal.NewApp(db)

	log.Info().Int("port", config.Cfg.AppPort).Msg("Starting server")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = app.Fiber.ShutdownWithContext(shutdownContext); err != nil {
		log.Error().Err(err).Msg("Error shutting down server")
	} else {
		log.Info().Msg("Server shutdown successfully")
	}
}
