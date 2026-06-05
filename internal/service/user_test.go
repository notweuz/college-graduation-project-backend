package service

import (
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserService_FindByID(t *testing.T) {
	mockUserDB := mocks.NewMockUserDatabase(t)

	exampleEmail := "test@example.com" // i dunno why I made email as pointer
	exampleFullName := "Anton"
	mockUserDB.On("FindByID", uint64(1)).Return(&model.User{
		ID:       1,
		Email:    &exampleEmail,
		FullName: &exampleFullName,
		Role:     enum.RoleAdmin,
	}, nil).Once()

	exampleEmail2 := "nastya@yandex.ru"
	exampleFullName2 := "Nastya"
	mockUserDB.On("FindByID", uint64(2)).Return(&model.User{
		ID:       2,
		Email:    &exampleEmail2,
		FullName: &exampleFullName2,
		Role:     enum.RoleClient,
	}, nil).Once()

	mockUserDB.On("FindByID", mock.AnythingOfType("uint64")).
		Return(nil, gorm.ErrRecordNotFound).
		Maybe()

	userSvcTest := NewUserService(mockUserDB)

	user, err := userSvcTest.FindByID(1)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, uint64(1))
	assert.Equal(t, *user.Email, exampleEmail)
	assert.Equal(t, *user.FullName, exampleFullName)
	assert.Equal(t, user.Role, enum.RoleAdmin)

	user, err = userSvcTest.FindByID(2)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, uint64(2))
	assert.Equal(t, *user.Email, exampleEmail2)
	assert.Equal(t, *user.FullName, exampleFullName2)
	assert.Equal(t, user.Role, enum.RoleClient)

	user, err = userSvcTest.FindByID(3)

	if err != nil {
		assert.Contains(t, err.Error(), "Cannot find user")
		assert.Nil(t, user)
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	mockUserDB := mocks.NewMockUserDatabase(t)

	id := uint64(1)
	email := "user@example.com"
	fullName := "User Example"
	mockUser := &model.User{
		ID:       id,
		Email:    &email,
		FullName: &fullName,
		Role:     enum.RoleAdmin,
	}
	mockUserDB.On("FindByID", uint64(1)).Return(mockUser, nil)
	mockUserDB.On("FindByEmail", "user2@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockUserDB.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	userSvcTest := NewUserService(mockUserDB)

	newMail := "user2@example.com"
	newFullName := "Ayanami Rei"
	updateProfile := request.UpdateProfile{
		Email:    &newMail,
		FullName: &newFullName,
	}
	mockUser, err := userSvcTest.UpdateProfile(uint64(1), updateProfile)

	assert.NoError(t, err)
	assert.Equal(t, mockUser.ID, uint64(1))
	assert.Equal(t, *mockUser.Email, newMail)
	assert.Equal(t, *mockUser.FullName, newFullName)
	assert.Equal(t, mockUser.Role, enum.RoleAdmin)
}

func TestUserService_UpdateProfile_EmailAlreadyExists(t *testing.T) {
	mockUserDB := mocks.NewMockUserDatabase(t)

	id := uint64(1)
	email := "user@example.com"
	fullName := "User Example"
	mockUser := &model.User{
		ID:       id,
		Email:    &email,
		FullName: &fullName,
		Role:     enum.RoleAdmin,
	}

	existingEmail := "busy@example.com"
	existingFullName := "Busy User"
	existingUser := &model.User{
		ID:       2,
		Email:    &existingEmail,
		FullName: &existingFullName,
		Role:     enum.RoleClient,
	}

	mockUserDB.On("FindByID", id).Return(mockUser, nil)
	mockUserDB.On("FindByEmail", existingEmail).Return(existingUser, nil)

	userSvcTest := NewUserService(mockUserDB)

	updateProfile := request.UpdateProfile{
		Email: &existingEmail,
	}
	updatedUser, err := userSvcTest.UpdateProfile(id, updateProfile)

	assert.Nil(t, updatedUser)
	var appErr *errs.AppError
	assert.ErrorAs(t, err, &appErr)
	assert.Equal(t, "user with this email already exists", appErr.Reason)
}

func TestUserService_FindByEmail(t *testing.T) {
	mockUserDB := mocks.NewMockUserDatabase(t)

	email := "xyz@xyz.xyz"
	fullName := "Test user"
	mockUser := &model.User{
		ID:           1,
		PasswordHash: "",
		Email:        &email,
		FullName:     &fullName,
	}

	mockUserDB.On("FindByEmail", "xyz@xyz.xyz").Return(mockUser, nil)
	mockUserDB.On("FindByEmail", mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound)

	userSvcTest := NewUserService(mockUserDB)

	user, err := userSvcTest.FindByEmail("xyz@xyz.xyz")

	assert.NoError(t, err)
	assert.Equal(t, user.ID, uint64(1))
	assert.Equal(t, *user.Email, email)
	assert.Equal(t, *user.FullName, fullName)

	user, err = userSvcTest.FindByEmail("notexist@notexist.notexist")

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, user)
}
