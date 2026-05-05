package service

import (
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupReview(t *testing.T) (*mocks.MockReviewDatabase, *mocks.MockUserService, *mocks.MockBookingService) {
	t.Helper()
	return mocks.NewMockReviewDatabase(t), mocks.NewMockUserService(t), mocks.NewMockBookingService(t)
}

func makeFinishedBooking(id, userID, hallID uint64) *model.Booking {
	return &model.Booking{
		ID:            id,
		UserID:        userID,
		HallID:        hallID,
		StartDateTime: time.Now().AddDate(0, 0, -5),
		EndDateTime:   time.Now().AddDate(0, 0, -1),
	}
}

func makeActiveBooking(id, userID, hallID uint64) *model.Booking {
	return &model.Booking{
		ID:            id,
		UserID:        userID,
		HallID:        hallID,
		StartDateTime: time.Now().AddDate(0, 0, 1),
		EndDateTime:   time.Now().AddDate(0, 0, 5),
	}
}

func TestReviewService_Create_Success(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	client := makeClientUser(1)
	booking := makeFinishedBooking(1, 1, 1)
	booking.User = *client

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	bookingSvc.On("FindByID", uint64(1), uint64(1)).Return(booking, nil)
	reviewDB.On("FindByBookingID", uint64(1)).Return(nil, gorm.ErrRecordNotFound)
	reviewDB.On("Create", mock.AnythingOfType("*model.Review")).Return(nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Create(1, 1, &request.ReviewCreate{Rating: 5, Comment: "Great!"})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint8(5), result.Rating)
}

func TestReviewService_Create_BookingNotFinished(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	client := makeClientUser(1)
	booking := makeActiveBooking(1, 1, 1)
	booking.User = *client

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	bookingSvc.On("FindByID", uint64(1), uint64(1)).Return(booking, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Create(1, 1, &request.ReviewCreate{Rating: 5, Comment: "Great!"})

	assert.Error(t, err)
	assert.Nil(t, result)
	reviewDB.AssertNotCalled(t, "Create")
}

func TestReviewService_Create_NotOwner(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	client := makeClientUser(2)
	booking := makeFinishedBooking(1, 1, 1)

	ownerEmail := "owner@example.com"
	ownerName := "Owner"
	booking.User = model.User{ID: 1, Email: &ownerEmail, FullName: &ownerName}

	userSvc.On("FindByID", uint64(2)).Return(client, nil)
	bookingSvc.On("FindByID", uint64(2), uint64(1)).Return(booking, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Create(2, 1, &request.ReviewCreate{Rating: 5, Comment: "Great!"})

	assert.Error(t, err)
	assert.Nil(t, result)
	reviewDB.AssertNotCalled(t, "Create")
}

func TestReviewService_Create_AlreadyReviewed(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	client := makeClientUser(1)
	booking := makeFinishedBooking(1, 1, 1)
	booking.User = *client

	existingReview := &model.Review{ID: 1, BookingID: 1, UserID: 1, Rating: 4}

	userSvc.On("FindByID", uint64(1)).Return(client, nil)
	bookingSvc.On("FindByID", uint64(1), uint64(1)).Return(booking, nil)
	reviewDB.On("FindByBookingID", uint64(1)).Return(existingReview, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Create(1, 1, &request.ReviewCreate{Rating: 5, Comment: "Again!"})

	assert.Error(t, err)
	assert.Nil(t, result)
	reviewDB.AssertNotCalled(t, "Create")
}

func TestReviewService_GetByBookingID_Success(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	client := makeClientUser(1)
	booking := makeFinishedBooking(1, 1, 1)
	booking.User = *client
	review := &model.Review{ID: 1, BookingID: 1, UserID: 1, Rating: 5}

	bookingSvc.On("FindByID", uint64(1), uint64(1)).Return(booking, nil)
	reviewDB.On("FindByBookingID", uint64(1)).Return(review, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.GetByBookingID(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, review, result)
}

func TestReviewService_GetByBookingID_NotFound(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	client := makeClientUser(1)
	booking := makeFinishedBooking(1, 1, 1)
	booking.User = *client

	bookingSvc.On("FindByID", uint64(1), uint64(1)).Return(booking, nil)
	reviewDB.On("FindByBookingID", uint64(1)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.GetByBookingID(1, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReviewService_GetByHallID(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	reviews := []model.Review{
		{ID: 1, BookingID: 1, UserID: 1, Rating: 5},
		{ID: 2, BookingID: 2, UserID: 2, Rating: 4},
	}

	reviewDB.On("FindByHallID", uint64(1)).Return(reviews, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.GetByHallID(1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestReviewService_GetAllWithFilters(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	hallID := uint64(1)
	minRating := uint64(4)
	reviews := []model.Review{
		{ID: 1, Rating: 4},
		{ID: 2, Rating: 5},
	}

	reviewDB.On("FindAllWithFilters", &hallID, &minRating).Return(reviews, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.GetAllWithFilters(&hallID, &minRating)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestReviewService_Update_Success(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	review := &model.Review{ID: 1, BookingID: 1, UserID: 1, Rating: 3, Comment: "Old comment"}

	reviewDB.On("FindByID", uint64(1)).Return(review, nil)
	reviewDB.On("Update", review).Return(nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Update(1, 1, &request.ReviewCreate{Rating: 5, Comment: "Updated comment"})

	assert.NoError(t, err)
	assert.Equal(t, uint8(5), result.Rating)
	assert.Equal(t, "Updated comment", result.Comment)
}

func TestReviewService_Update_NotOwner(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	review := &model.Review{ID: 1, BookingID: 1, UserID: 1, Rating: 3}

	reviewDB.On("FindByID", uint64(1)).Return(review, nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Update(2, 1, &request.ReviewCreate{Rating: 5, Comment: "Updated"})

	assert.Error(t, err)
	assert.Nil(t, result)
	reviewDB.AssertNotCalled(t, "Update")
}

func TestReviewService_Update_NotFound(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	reviewDB.On("FindByID", uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	result, err := svc.Update(1, 99, &request.ReviewCreate{Rating: 5, Comment: "Updated"})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestReviewService_Delete(t *testing.T) {
	reviewDB, userSvc, bookingSvc := setupReview(t)

	reviewDB.On("Delete", uint64(1)).Return(nil)

	svc := NewReviewService(reviewDB, bookingSvc, userSvc)

	err := svc.Delete(1)

	assert.NoError(t, err)
}
