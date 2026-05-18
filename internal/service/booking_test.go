package service

import (
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupBooking(t *testing.T) (*mocks.MockBookingDatabase, *mocks.MockUserService, *mocks.MockHallService) {
	t.Helper()
	return mocks.NewMockBookingDatabase(t), mocks.NewMockUserService(t), mocks.NewMockHallService(t)
}

func futureDate(daysFromNow int) string {
	return time.Now().AddDate(0, 0, daysFromNow).Format("2006-01-02")
}

func futureDatetime(daysFromNow int) time.Time {
	t := time.Now().AddDate(0, 0, daysFromNow)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func makeBooking(id, userID, hallID uint64, start, end time.Time) *model.Booking {
	return &model.Booking{
		ID:            id,
		UserID:        userID,
		HallID:        hallID,
		StartDateTime: start,
		EndDateTime:   end,
		TotalPrice:    100.0,
	}
}

func TestBookingService_Create_Success(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)
	bookingDB.On("CheckConflict", uint64(1), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(false, nil)
	bookingDB.On("Create", mock.AnythingOfType("*model.Booking")).Return(nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.Create(1, &request.BookingCreate{
		HallID:        1,
		StartDateTime: futureDate(2),
		EndDateTime:   futureDate(3),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestBookingService_Create_HallNotActive(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	hall := makeHall(1, "Hall A", false)

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.Create(1, &request.BookingCreate{
		HallID:        1,
		StartDateTime: futureDate(2),
		EndDateTime:   futureDate(3),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	bookingDB.AssertNotCalled(t, "Create")
}

func TestBookingService_Create_StartInPast(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.Create(1, &request.BookingCreate{
		HallID:        1,
		StartDateTime: futureDate(-5),
		EndDateTime:   futureDate(-1),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	bookingDB.AssertNotCalled(t, "Create")
}

func TestBookingService_Create_Conflict(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)
	bookingDB.On("CheckConflict", uint64(1), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(true, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.Create(1, &request.BookingCreate{
		HallID:        1,
		StartDateTime: futureDate(2),
		EndDateTime:   futureDate(3),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	bookingDB.AssertNotCalled(t, "Create")
}

func TestBookingService_Create_InvalidDateFormat(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.Create(1, &request.BookingCreate{
		HallID:        1,
		StartDateTime: "not-a-date",
		EndDateTime:   futureDate(3),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestBookingService_FindByID_OwnBooking(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	client.ID = 1
	booking := makeBooking(1, 1, 1, futureDatetime(2), futureDatetime(3))
	booking.User = *client

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.FindByID(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, booking, result)
}

func TestBookingService_FindByID_NotOwner(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(2)
	booking := makeBooking(1, 1, 1, futureDatetime(2), futureDatetime(3))

	ownerEmail := "owner@example.com"
	ownerName := "Owner"
	booking.User = model.User{ID: 1, Email: &ownerEmail, FullName: &ownerName, Role: enum.RoleClient}

	userSvc.On("FindByID", uint64(2)).Return(client, nil)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.FindByID(2, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestBookingService_FindByID_AdminCanSeeAll(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	admin := makeAdminUser(1)
	booking := makeBooking(1, 2, 1, futureDatetime(2), futureDatetime(3))

	otherEmail := "other@example.com"
	otherName := "Other"
	booking.User = model.User{ID: 2, Email: &otherEmail, FullName: &otherName, Role: enum.RoleClient}

	userSvc.On("FindByID", uint64(1)).Return(admin, nil)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.FindByID(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, booking, result)
}

func TestBookingService_FindAllFromUser_Success(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	bookings := []model.Booking{
		*makeBooking(1, 1, 1, futureDatetime(2), futureDatetime(3)),
	}

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	bookingDB.On("FindAllFromUser", uint64(1), (*time.Time)(nil), (*time.Time)(nil)).Return(bookings, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.FindAllFromUser(1, nil, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestBookingService_DeleteByAuthor_Success(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	booking := makeBooking(1, 1, 1, futureDatetime(5), futureDatetime(7))
	booking.User = *client

	userSvc.On("FindByID", uint64(1)).Return(client, nil).Times(2)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)
	bookingDB.On("Delete", uint64(1)).Return(nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	err := svc.DeleteByAuthor(1, 1)

	assert.NoError(t, err)
}

func TestBookingService_DeleteByAuthor_AlreadyStarted(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(1)
	booking := makeBooking(1, 1, 1, futureDatetime(-3), futureDatetime(-1))
	booking.User = *client

	userSvc.On("FindByID", uint64(1)).Return(client, nil).Times(2)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	err := svc.DeleteByAuthor(1, 1)

	assert.Error(t, err)
	bookingDB.AssertNotCalled(t, "Delete")
}

func TestBookingService_FindAll_AdminOnly(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	admin := makeAdminUser(1)
	bookings := []model.Booking{
		*makeBooking(1, 2, 1, futureDatetime(2), futureDatetime(3)),
	}

	userSvc.On("FindByID", uint64(1)).Return(admin, nil)
	bookingDB.On("FindAll", (*uint64)(nil), (*uint64)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(bookings, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.FindAll(1, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestBookingService_FindAll_NotAdmin(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(2)
	userSvc.On("FindByID", uint64(2)).Return(client, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	result, err := svc.FindAll(2, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	bookingDB.AssertNotCalled(t, "FindAll")
}

func TestBookingService_Update_AdminSuccess(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	admin := makeAdminUser(1)
	booking := makeBooking(1, 2, 1, futureDatetime(2), futureDatetime(3))
	booking.User = *admin

	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Times(2)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)
	bookingDB.On("Update", booking).Return(nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	newComment := "Updated comment"
	result, err := svc.Update(1, 1, &request.BookingUpdate{Comment: &newComment})

	assert.NoError(t, err)
	assert.Equal(t, newComment, result.Comment)
}

func TestBookingService_Update_NotAdmin(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	client := makeClientUser(2)
	userSvc.On("FindByID", uint64(2)).Return(client, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	newComment := "Updated comment"
	result, err := svc.Update(2, 1, &request.BookingUpdate{Comment: &newComment})

	assert.Error(t, err)
	assert.Nil(t, result)
	bookingDB.AssertNotCalled(t, "Update")
}

func TestBookingService_CalculatePrice_OneDay(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	hall := makeHall(1, "Hall A", true)
	hall.PricePerDay = 100.0
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	from := futureDatetime(2)
	to := futureDatetime(2)

	result, err := svc.CalculatePrice(1, from, to)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100.0, result.Price)
	assert.Zero(t, result.Discount)
	assert.Equal(t, 100.0, result.Total)
}

func TestBookingService_CalculatePrice_MultiDay_Discount(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	hall := makeHall(1, "Hall A", true)
	hall.PricePerDay = 100.0
	hallSvc.On("FindByID", uint64(1)).Return(hall, nil)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	from := futureDatetime(2)
	to := futureDatetime(4)

	result, err := svc.CalculatePrice(1, from, to)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 300.0, result.Price)
	assert.Equal(t, -90.0, result.Discount)
	assert.NotZero(t, result.Discount)
	assert.Equal(t, 210.0, result.Total)
}

func TestBookingService_Update_RecordNotFound(t *testing.T) {
	bookingDB, userSvc, hallSvc := setupBooking(t)

	admin := makeAdminUser(1)
	booking := makeBooking(1, 1, 1, futureDatetime(2), futureDatetime(3))
	booking.User = *admin

	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Times(2)
	bookingDB.On("FindByID", uint64(1)).Return(booking, nil)
	bookingDB.On("Update", booking).Return(gorm.ErrRecordNotFound)

	svc := NewBookingService(bookingDB, userSvc, hallSvc)

	newComment := "comment"
	result, err := svc.Update(1, 1, &request.BookingUpdate{Comment: &newComment})

	assert.Error(t, err)
	assert.Nil(t, result)
}
