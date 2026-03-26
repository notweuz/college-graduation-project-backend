package service

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userService UserService
}

func NewAuthService(userService UserService) AuthService {
	return &authService{userService: userService}
}

func (s *authService) Register(req *request.Register) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errs.InternalServerError("Cannot register user", "failed to hash password")
	}

	email := req.Email
	fullName := req.FullName

	user := &model.User{
		Email:        &email,
		FullName:     &fullName,
		PasswordHash: string(hash),
	}

	if err := s.userService.Create(user); err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	log.Info().Uint64("id", user.ID).Str("email", *user.Email).Msg("User successfully registered")

	signedToken, err := token.SignedString([]byte(config.Cfg.JwtSecret))
	if err != nil {
		return "", errs.InternalServerError("Cannot register user", "failed to generate access token")
	}
	return signedToken, nil
}

func (s *authService) Login(req *request.Login) (string, error) {
	user, err := s.userService.FindByEmail(req.Email)
	if err != nil {
		return "", errs.Unauthorized("Authentication failed", "invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errs.Unauthorized("Authentication failed", "invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	log.Info().Uint64("id", user.ID).Str("email", *user.Email).Msg("User successfully logged in")

	signedToken, err := token.SignedString([]byte(config.Cfg.JwtSecret))
	if err != nil {
		return "", errs.InternalServerError("Cannot login", "failed to generate access token")
	}

	return signedToken, nil
}
