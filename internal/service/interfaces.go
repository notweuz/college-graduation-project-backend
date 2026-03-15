package service

import "college-graduation-project-backend/internal/model"

type UserService interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
}
