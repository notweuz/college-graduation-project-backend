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
	FindByHallID(hallID uint64) ([]model.Review, error)
	FindByBookingID(bookingID uint64) (*model.Review, error)
	FindByUserIDAndBookingID(userID, bookingID uint64) (*model.Review, error)
	FindAllWithFilters(hallID, minRating *uint64) ([]model.Review, error)
	Update(review *model.Review) error
	Delete(id uint64) error
}

type ImageDatabase interface {
	Create(image *model.Image) error
	GetByID(id uint64) (*model.Image, error)
	FindByHallID(hallID uint64) ([]model.Image, error)
	GetByUserID(userID uint64) (*model.Image, error)
	AttachToHall(hallID, imageID uint64) error
	SetUserImage(userID, imageID uint64) error
	RemoveUserImage(userID uint64) error
	Delete(id uint64) error
}

type BookingDatabase interface {
	Create(booking *model.Booking) error
	FindByID(id uint64) (*model.Booking, error)
	FindAllFromUser(userID uint64, from, to *time.Time) ([]model.Booking, error)
	Update(booking *model.Booking) error
	Delete(id uint64) error
	CheckConflict(hallID uint64, startDateTime, endDateTime time.Time) (bool, error)
	FindBookingsForHall(hallID uint64, from, to time.Time) ([]model.Booking, error)
	FindAll(hallID, bookingUserID *uint64, from, to *time.Time) ([]model.Booking, error)
}

type ReportDatabase interface {
	FindSalesBookings(from, to time.Time, hallID *uint64) ([]model.Booking, error)
	FindReportBookings(from, to time.Time, hallID *uint64) ([]model.Booking, error)
	CountHalls(hallID *uint64) (uint64, error)
}

type UserAgreementDatabase interface {
	Get() (*model.UserAgreement, error)
	Save(text string) (*model.UserAgreement, error)
}
