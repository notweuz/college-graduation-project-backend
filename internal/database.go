package internal

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/model"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupDatabase() (*gorm.DB, error) {
	dsn := config.Cfg.DatabaseDSN
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Connected to database")

	err = db.AutoMigrate(
		&model.User{},
		&model.Hall{},
		&model.Image{},
		&model.HallImage{},
		&model.UserImage{},
		&model.Booking{},
		&model.Review{},
		&model.UserAgreement{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	log.Info().Msg("Migrated database successfully")
	return db, err
}
