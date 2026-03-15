package service

import (
	"college-graduation-project-backend/internal/config"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		return "", err
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
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(config.Cfg.JwtSecret))
}

func (s *authService) Login(req *request.Login) (string, error) {
	user, err := s.userService.FindByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(config.Cfg.JwtSecret))
}
