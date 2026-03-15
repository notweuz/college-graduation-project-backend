package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	DatabaseDSN string `envconfig:"DATABASE_DSN" default:""`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	JwtSecret   string `envconfig:"JWT_SECRET" default:""`
	AppPort     int    `envconfig:"APP_PORT" default:"8080"`
}

var Cfg Config

func SetupConfig() error {
	_ = godotenv.Load()
	if err := envconfig.Process("", &Cfg); err != nil {
		return err
	}
	log.Info().Msg("Config loaded")

	level, err := zerolog.ParseLevel(Cfg.LogLevel)
	if err != nil {
		log.Error().Str("originalLogLevel", Cfg.LogLevel).Msg("Failed to parse log level, fallback to info level. Maybe a typo?")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Info().Msgf("Set log level to %s", level.String())

	return nil
}
