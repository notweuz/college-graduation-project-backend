package service

import (
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
	"college-graduation-project-backend/internal/model/response"
	"time"
)

type UserService interface {
	Create(user *model.User) error
	FindByID(id uint64) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	Delete(id uint64) error
	UpdateProfile(userID uint64, updateProfile request.UpdateProfile) (*model.User, error)
}

type AuthService interface {
	Register(req *request.Register) (string, error)
	Login(req *request.Login) (string, error)
}

type HallService interface {
	Create(userID uint64, hallCreate *request.HallCreate) (*model.Hall, error)
	FindByID(id uint64) (*model.Hall, error)
	FindAll() ([]model.Hall, error)
	FindAllActive() ([]model.Hall, error)
	Update(userID, id uint64, hallUpdate *request.HallUpdate) (*model.Hall, error)
	Delete(userID, id uint64) error
	GetAvailability(hallID uint64, from, to time.Time) ([]response.HallAvailability, error)
	UploadImage(userID, hallID uint64, imagePath string) error
	GetHallImages(hallID uint64) ([]string, error)
}

type BookingService interface {
	Create(userID uint64, req *request.BookingCreate) (*model.Booking, error)
	FindAllFromUser(userID uint64, from, to *time.Time) ([]model.Booking, error)
	FindByID(userID, id uint64) (*model.Booking, error)
	DeleteByAuthor(userID, id uint64) error
	FindAll(userID uint64, hallID, bookingUserID *uint64, from, to *time.Time) ([]model.Booking, error)
	Update(userID uint64, id uint64, req *request.BookingUpdate) (*model.Booking, error)
}

type ReviewService interface {
	Create(userID, bookingID uint64, req *request.ReviewCreate) (*model.Review, error)
	GetByHallID(hallID uint64) ([]model.Review, error)
	GetAllWithFilters(hallID, minRating *uint64) ([]model.Review, error)
	Update(userID, reviewID uint64, req *request.ReviewCreate) (*model.Review, error)
	Delete(id uint64) error
}

type ReportService interface {
	GetSalesReport(userID uint64, from, to time.Time, hallID *uint64, groupBy string, metrics []string) (*response.SalesReport, error)
	GetSalesReportPDF(userID uint64, from, to time.Time, hallID *uint64, groupBy string, metrics []string) ([]byte, error)
	GetHallsLoadReport(userID uint64, from, to time.Time, hallID *uint64) (*response.HallsLoadReport, error)
	GetClientsReport(userID uint64, from, to time.Time, hallID *uint64, limit uint64) (*response.ClientsReport, error)
	GetBookingsDynamicsReport(userID uint64, from, to time.Time, hallID *uint64, groupBy string) (*response.BookingsDynamicsReport, error)
	GetHallsLoadReportPDF(userID uint64, from, to time.Time, hallID *uint64) ([]byte, error)
	GetClientsReportPDF(userID uint64, from, to time.Time, hallID *uint64, limit uint64) ([]byte, error)
	GetBookingsDynamicsReportPDF(userID uint64, from, to time.Time, hallID *uint64, groupBy string) ([]byte, error)
}
