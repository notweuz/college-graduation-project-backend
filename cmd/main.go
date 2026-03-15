package main

import (
	"college-graduation-project-backend/internal"
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/logger"

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
	reviewDatabase := database.NewReviewDatabase(internal.Database)
	bookingDatabase := database.NewBookingDatabase(internal.Database)
}
