package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseDSN string `envconfig:"DATABASE_DSN" default:""`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	JwtSecret   string `envconfig:"JWT_SECRET" default:""`
	AppPort     int    `envconfig:"APP_PORT" default:"8080"`
	AppTimeZone string `envconfig:"APP_TIME_ZONE" default:"UTC"`
}

var Cfg Config
var AppLocation = time.UTC

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

	loc, err := parseAppTimeZone(Cfg.AppTimeZone)
	if err != nil {
		log.Error().Err(err).Str("originalTimeZone", Cfg.AppTimeZone).Msg("Failed to parse APP_TIME_ZONE, fallback to UTC")
		loc = time.UTC
	}
	AppLocation = loc
	log.Info().Str("timeZone", AppLocation.String()).Msg("Set app timezone")

	return nil
}

func parseAppTimeZone(value string) (*time.Location, error) {
	tz := strings.TrimSpace(value)
	if tz == "" {
		return time.UTC, nil
	}

	if loc, err := time.LoadLocation(tz); err == nil {
		return loc, nil
	}

	upper := strings.ToUpper(tz)
	if !strings.HasPrefix(upper, "UTC") {
		return nil, fmt.Errorf("unsupported timezone format: %s", tz)
	}

	offsetPart := strings.TrimSpace(tz[3:])
	if offsetPart == "" {
		return time.UTC, nil
	}

	sign := 1
	switch offsetPart[0] {
	case '+':
		sign = 1
	case '-':
		sign = -1
	default:
		return nil, fmt.Errorf("unsupported UTC offset format: %s", tz)
	}

	hours, err := strconv.Atoi(strings.TrimSpace(offsetPart[1:]))
	if err != nil || hours > 14 {
		return nil, fmt.Errorf("invalid UTC offset hours: %s", tz)
	}

	return time.FixedZone("UTC"+offsetPart, sign*hours*3600), nil
}
