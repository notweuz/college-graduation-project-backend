package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func Logging() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		log.Debug().Msgf("HTTP %s %s from %s", c.Method(), c.Path(), c.IP())

		err := c.Next()
		duration := time.Since(start)
		status := c.Response().StatusCode()

		if err != nil {
			log.Debug().Msgf("HTTP %s %s failed in %v", c.Method(), c.Path(), duration)
			return err
		}

		log.Info().Msgf("HTTP %s %s completed with status %d in %v", c.Method(), c.Path(), status, duration)
		return nil
	}
}
