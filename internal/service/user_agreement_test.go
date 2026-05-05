package service

import (
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) *mocks.MockUserAgreementDatabase {
	t.Helper()
	return mocks.NewMockUserAgreementDatabase(t)
}

func TestUserAgreementService_Get(t *testing.T) {
	mockUserAgreementDB := setup(t)
	mockUserSvc := mocks.NewMockUserService(t)

	exampleAgreement := &model.UserAgreement{
		ID:        1,
		Text:      "User Agreement Example 1234",
		Version:   1,
		UpdatedAt: time.Now(),
	}

	mockUserAgreementDB.On("Get").
		Return(exampleAgreement, nil).
		Once()

	userAgreementSvcTest := NewUserAgreementService(mockUserAgreementDB, mockUserSvc)

	agreement, err := userAgreementSvcTest.Get()

	assert.NoError(t, err)
	assert.Equal(t, exampleAgreement, agreement)
	mockUserSvc.AssertNotCalled(t, "FindByID")
}

func TestUserAgreementService_Update(t *testing.T) {
	mockUserAgreementDB := setup(t)
	mockUserSvc := mocks.NewMockUserService(t)

	exampleAgreement := &model.UserAgreement{
		ID:        1,
		Text:      "User Agreement Example 1234",
		Version:   1,
		UpdatedAt: time.Now(),
	}

	updatedAgreement := &model.UserAgreement{
		ID:        1,
		Text:      "test update",
		Version:   2,
		UpdatedAt: time.Now(),
	}

	adminID := uint64(1)
	adminEmail := "admin@example.com"
	adminFullName := "Admin Example"
	adminMockUser := &model.User{
		ID:       adminID,
		Email:    &adminEmail,
		FullName: &adminFullName,
		Role:     enum.RoleAdmin,
	}

	userID := uint64(2)
	userEmail := "user@example.com"
	userFullName := "User Example"
	userMockUser := &model.User{
		ID:       userID,
		Email:    &userEmail,
		FullName: &userFullName,
		Role:     enum.RoleClient,
	}

	mockUserAgreementDB.On("Get").
		Return(exampleAgreement, nil).
		Once()

	mockUserAgreementDB.On("Save", "test update").
		Return(updatedAgreement, nil).
		Once()

	mockUserSvc.On("FindByID", uint64(1)).
		Return(adminMockUser, nil)

	mockUserSvc.On("FindByID", uint64(2)).
		Return(userMockUser, nil)

	userAgreementSvcTest := NewUserAgreementService(mockUserAgreementDB, mockUserSvc)

	agreement, err := userAgreementSvcTest.Get()
	assert.NoError(t, err)
	assert.Equal(t, exampleAgreement, agreement)

	result, err := userAgreementSvcTest.Update(userID, "test update")
	assert.Error(t, err)
	assert.Nil(t, result)
	mockUserAgreementDB.AssertNotCalled(t, "Save")

	result, err = userAgreementSvcTest.Update(adminID, "test update")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, updatedAgreement.Text, result.Text)
	assert.Equal(t, updatedAgreement.Version, result.Version)
}
