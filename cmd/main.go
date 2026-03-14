package main

import (
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

	err = database.SetupDatabase()
	if err != nil {
		log.Panic().Err(err).Msg("Error loading database")
	}
}
