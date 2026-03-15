package database

import "college-graduation-project-backend/internal/model"

type UserDatabase interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
}

type HallDatabase interface {
	Create(hall *model.Hall) error
	FindByID(id uint) (*model.Hall, error)
	FindAll() ([]model.Hall, error)
	Update(hall *model.Hall) error
	Delete(id uint) error
}

type ReviewDatabase interface {
	Create(review *model.Review) error
	FindByID(id uint) (*model.Review, error)
	FindAll() ([]model.Review, error)
	Update(review *model.Review) error
	Delete(id uint) error
}

type BookingDatabase interface {
	Create(booking *model.Booking) error
	FindByID(id uint) (*model.Booking, error)
	FindAll() ([]model.Booking, error)
	Update(booking *model.Booking) error
	Delete(id uint) error
}
