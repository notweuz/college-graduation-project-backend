package service

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupHall(t *testing.T) (*mocks.MockHallDatabase, *mocks.MockBookingDatabase, *mocks.MockUserService) {
	t.Helper()
	return mocks.NewMockHallDatabase(t), mocks.NewMockBookingDatabase(t), mocks.NewMockUserService(t)
}

func makeAdminUser(id uint64) *model.User {
	return makeUser(id, "admin@example.com", "Admin User", enum.RoleAdmin, "hash")
}

func makeClientUser(id uint64) *model.User {
	return makeUser(id, "client@example.com", "Client User", enum.RoleClient, "hash")
}

func makeHall(id uint64, name string, active bool) *model.Hall {
	return model.NewHall(id, name, "description", 100.0, active)
}

func TestHallService_Create_Success(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()
	hallDB.On("Create", mock.AnythingOfType("*model.Hall")).Return(nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Create(1, &request.HallCreate{
		Name:        hall.Name,
		Description: hall.Description,
		PricePerDay: &hall.PricePerDay,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestHallService_Create_NotAdmin(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	client := makeClientUser(2)
	userSvc.On("FindByID", uint64(2)).Return(client, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Create(2, &request.HallCreate{Name: "Hall B", PricePerDay: ptr(100.0)})

	assert.Error(t, err)
	assert.Nil(t, result)
	hallDB.AssertNotCalled(t, "Create")
}

func TestHallService_Create_Duplicate(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()
	hallDB.On("Create", mock.AnythingOfType("*model.Hall")).Return(gorm.ErrDuplicatedKey).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Create(1, &request.HallCreate{Name: "Hall A", PricePerDay: ptr(100.0)})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHallService_Create_WithoutPrice(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Create(1, &request.HallCreate{Name: "Hall A"})

	assert.Nil(t, result)
	var appErr *errs.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "price_per_day must be greater than 0", appErr.Reason)
	hallDB.AssertNotCalled(t, "Create")
}

func TestHallService_Create_NegativePrice(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Create(1, &request.HallCreate{Name: "Hall A", PricePerDay: ptr(-1.0)})

	assert.Nil(t, result)
	var appErr *errs.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "price_per_day must be greater than 0", appErr.Reason)
	hallDB.AssertNotCalled(t, "Create")
}

func TestHallService_FindByID_Found(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	hall := makeHall(1, "Hall A", true)
	hallDB.On("FindByID", uint64(1)).Return(hall, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.FindByID(1)

	assert.NoError(t, err)
	assert.Equal(t, hall, result)
}

func TestHallService_FindByID_NotFound(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	hallDB.On("FindByID", uint64(99)).Return(nil, gorm.ErrRecordNotFound).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.FindByID(99)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHallService_FindAll(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	halls := []model.Hall{*makeHall(1, "Hall A", true), *makeHall(2, "Hall B", false)}
	hallDB.On("FindAll").Return(halls, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.FindAll()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestHallService_FindAllActive(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	halls := []model.Hall{*makeHall(1, "Hall A", true)}
	hallDB.On("FindAllActive").Return(halls, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.FindAllActive()

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestHallService_Update_Success(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()
	hallDB.On("FindByID", uint64(1)).Return(hall, nil).Once()
	hallDB.On("Update", hall).Return(nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	newName := "Hall A Updated"
	result, err := svc.Update(1, 1, &request.HallUpdate{Name: &newName})

	assert.NoError(t, err)
	assert.Equal(t, newName, result.Name)
}

func TestHallService_Update_NotAdmin(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	client := makeClientUser(2)
	userSvc.On("FindByID", uint64(2)).Return(client, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	newName := "Hall A Updated"
	result, err := svc.Update(2, 1, &request.HallUpdate{Name: &newName})

	assert.Error(t, err)
	assert.Nil(t, result)
	hallDB.AssertNotCalled(t, "Update")
}

func TestHallService_Update_NegativePrice(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()
	hallDB.On("FindByID", uint64(1)).Return(hall, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Update(1, 1, &request.HallUpdate{PricePerDay: ptr(-1.0)})

	assert.Nil(t, result)
	var appErr *errs.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "price_per_day must be greater than 0", appErr.Reason)
	hallDB.AssertNotCalled(t, "Update")
}

func TestHallService_Delete_Success(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()
	hallDB.On("Delete", uint64(1)).Return(nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	err := svc.Delete(1, 1)

	assert.NoError(t, err)
}

func TestHallService_Delete_NotAdmin(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	client := makeClientUser(2)
	userSvc.On("FindByID", uint64(2)).Return(client, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	err := svc.Delete(2, 1)

	assert.Error(t, err)
	hallDB.AssertNotCalled(t, "Delete")
}

func TestHallService_GetAvailability_NoBookings(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	hall := makeHall(1, "Hall A", true)
	from := time.Now()
	to := from.Add(48 * time.Hour)

	hallDB.On("FindByID", uint64(1)).Return(hall, nil).Once()
	bookingDB.On("FindBookingsForHall", uint64(1), from, to).Return([]model.Booking{}, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.GetAvailability(1, from, to)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "free", result[0].Status)
}

func TestHallService_GetAvailability_WithBookings(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	hall := makeHall(1, "Hall A", true)
	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 6, 10, 23, 59, 59, 0, time.UTC)

	bookingStart := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	bookingEnd := time.Date(2025, 6, 5, 23, 59, 59, 0, time.UTC)

	bookings := []model.Booking{
		{ID: 1, HallID: 1, StartDateTime: bookingStart, EndDateTime: bookingEnd},
	}

	hallDB.On("FindByID", uint64(1)).Return(hall, nil).Once()
	bookingDB.On("FindBookingsForHall", uint64(1), from, to).Return(bookings, nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.GetAvailability(1, from, to)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)

	statuses := make([]string, len(result))
	for i, r := range result {
		statuses[i] = r.Status
	}
	assert.Contains(t, statuses, "busy")
	assert.Contains(t, statuses, "free")
}

func TestHallService_GetAvailability_HallNotFound(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	hallDB.On("FindByID", uint64(99)).Return(nil, gorm.ErrRecordNotFound).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.GetAvailability(99, time.Now(), time.Now().Add(24*time.Hour))

	assert.Error(t, err)
	assert.Nil(t, result)
	bookingDB.AssertNotCalled(t, "FindBookingsForHall")
}

func ptr[T any](v T) *T { return &v }

func TestHallService_Update_AllFields(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	admin := makeAdminUser(1)
	hall := makeHall(1, "Hall A", true)

	userSvc.On("FindByID", uint64(1)).Return(admin, nil).Once()
	hallDB.On("FindByID", uint64(1)).Return(hall, nil).Once()
	hallDB.On("Update", mock.AnythingOfType("*model.Hall")).Return(nil).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.Update(1, 1, &request.HallUpdate{
		Name:        ptr("New Name"),
		Description: ptr("New Desc"),
		PricePerDay: ptr(200.0),
		IsActive:    ptr(false),
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, "New Desc", result.Description)
	assert.Equal(t, 200.0, result.PricePerDay)
	assert.False(t, result.IsActive)
}

func TestHallService_FindAll_DBError(t *testing.T) {
	hallDB, bookingDB, userSvc := setupHall(t)

	hallDB.On("FindAll").Return(nil, errors.New("db error")).Once()

	svc := NewHallService(hallDB, bookingDB, userSvc)

	result, err := svc.FindAll()

	assert.Error(t, err)
	assert.Nil(t, result)
}
