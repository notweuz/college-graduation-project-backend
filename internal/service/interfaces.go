package service

import (
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
)

type UserService interface {
	Create(user *model.User) error
	FindByID(id uint64) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	Delete(id uint64) error
}

type AuthService interface {
	Register(req *request.Register) (string, error)
	Login(req *request.Login) (string, error)
}
