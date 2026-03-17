package service

import (
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"

	"github.com/gofiber/fiber/v3"
)

type UserService interface {
	Create(user *model.User) error
	FindByID(id uint64) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	Delete(id uint64) error
	UpdateProfile(ctx fiber.Ctx, updateProfile request.UpdateProfile) (*model.User, error)
}

type AuthService interface {
	Register(req *request.Register) (string, error)
	Login(req *request.Login) (string, error)
}

type HallService interface {
	Create(ctx fiber.Ctx, hallCreate *request.HallCreate) (*model.Hall, error)
	FindByID(id uint64) (*model.Hall, error)
	FindAll() ([]model.Hall, error)
	FindAllActive() ([]model.Hall, error)
	Update(ctx fiber.Ctx, hallUpdate *request.HallUpdate) (*model.Hall, error)
	Delete(id uint64) error
}
