package service

import (
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
	mockUserDatabase := mocks.NewMockUserDatabase(t)

	exampleEmail := "test@example.com" // i dunno why I made email as pointer
	exampleFullName := "Anton"
	mockUserDatabase.On("FindByID", uint64(1)).Return(&model.User{
		ID:       1,
		Email:    &exampleEmail,
		FullName: &exampleFullName,
		Role:     enum.RoleAdmin,
	}, nil).Once()

	exampleEmail2 := "nastya@yandex.ru"
	exampleFullName2 := "Nastya"
	mockUserDatabase.On("FindByID", uint64(2)).Return(&model.User{
		ID:       2,
		Email:    &exampleEmail2,
		FullName: &exampleFullName2,
		Role:     enum.RoleClient,
	}, nil).Once()

	mockUserDatabase.On("FindByID", mock.AnythingOfType("uint64")).
		Return(nil, gorm.ErrRecordNotFound).
		Maybe()

	userServiceTest := NewUserService(mockUserDatabase)

	user, err := userServiceTest.FindByID(1)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, uint64(1))
	assert.Equal(t, *user.Email, exampleEmail)
	assert.Equal(t, *user.FullName, exampleFullName)
	assert.Equal(t, user.Role, enum.RoleAdmin)

	user, err = userServiceTest.FindByID(2)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, uint64(2))
	assert.Equal(t, *user.Email, exampleEmail2)
	assert.Equal(t, *user.FullName, exampleFullName2)
	assert.Equal(t, user.Role, enum.RoleClient)

	user, err = userServiceTest.FindByID(3)

	if err != nil {
		assert.Contains(t, err.Error(), "Cannot find user")
		assert.Nil(t, user)
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	mockUserDatabase := mocks.NewMockUserDatabase(t)

	id := uint64(1)
	email := "user@example.com"
	fullName := "User Example"
	mockUser := &model.User{
		ID:       id,
		Email:    &email,
		FullName: &fullName,
		Role:     enum.RoleAdmin,
	}
	mockUserDatabase.On("FindByID", uint64(1)).Return(mockUser, nil)
	mockUserDatabase.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	userServiceTest := NewUserService(mockUserDatabase)

	newMail := "user2@example.com"
	newFullName := "Ayanami Rei"
	updateProfile := request.UpdateProfile{
		Email:    &newMail,
		FullName: &newFullName,
	}
	mockUser, err := userServiceTest.UpdateProfile(uint64(1), updateProfile)

	assert.NoError(t, err)
	assert.Equal(t, mockUser.ID, uint64(1))
	assert.Equal(t, *mockUser.Email, newMail)
	assert.Equal(t, *mockUser.FullName, newFullName)
	assert.Equal(t, mockUser.Role, enum.RoleAdmin)
}

func TestUserService_FindByEmail(t *testing.T) {
	mockUserDatabase := mocks.NewMockUserDatabase(t)

	email := "xyz@xyz.xyz"
	fullName := "Test user"
	mockUser := &model.User{
		ID:           1,
		PasswordHash: "",
		Email:        &email,
		FullName:     &fullName,
	}

	mockUserDatabase.On("FindByEmail", "xyz@xyz.xyz").Return(mockUser, nil)
	mockUserDatabase.On("FindByEmail", mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound)

	userServiceTest := NewUserService(mockUserDatabase)

	user, err := userServiceTest.FindByEmail("xyz@xyz.xyz")

	assert.NoError(t, err)
	assert.Equal(t, user.ID, uint64(1))
	assert.Equal(t, *user.Email, email)
	assert.Equal(t, *user.FullName, fullName)

	user, err = userServiceTest.FindByEmail("notexist@notexist.notexist")

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, user)
}
