package middleware

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/errs"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

func Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return errs.Unauthorized("Not authorized", "Missing Bearer token")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.Cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			return err
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Locals("user_id", claims["user_id"])
		c.Locals("role", claims["role"])

		return c.Next()
	}
}

func GetCurrentUserID(c fiber.Ctx) (uint64, error) {
	userID, ok := c.Locals("user_id").(float64)
	if !ok {
		log.Error().Msg("Cannot get current user id")
		return 0, errs.Unauthorized("Not authorized", "user_id not found in locals")
	}
	return uint64(userID), nil
}
