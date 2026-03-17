package database

import (
	"college-graduation-project-backend/internal/model"
	"time"
)

type UserDatabase interface {
	Create(user *model.User) error
	FindByID(id uint64) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	Delete(id uint64) error
}

type HallDatabase interface {
	Create(hall *model.Hall) error
	FindByID(id uint64) (*model.Hall, error)
	FindAll() ([]model.Hall, error)
	FindAllActive() ([]model.Hall, error)
	Update(hall *model.Hall) error
	Delete(id uint64) error
}

type ReviewDatabase interface {
	Create(review *model.Review) error
	FindByID(id uint64) (*model.Review, error)
	FindAll() ([]model.Review, error)
	Update(review *model.Review) error
	Delete(id uint64) error
}

type BookingDatabase interface {
	Create(booking *model.Booking) error
	FindByID(id uint64) (*model.Booking, error)
	FindAll() ([]model.Booking, error)
	FindAllFromUser(userID uint64, from, to *time.Time) ([]model.Booking, error)
	Update(booking *model.Booking) error
	Delete(id uint64) error
}
