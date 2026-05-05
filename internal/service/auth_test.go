package service

import (
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupAuth(t *testing.T) *mocks.MockUserService {
	t.Helper()
	return mocks.NewMockUserService(t)
}

func makeUser(id uint64, email, fullName string, role enum.UserRole, passwordHash string) *model.User {
	e := email
	fn := fullName
	return &model.User{
		ID:           id,
		Email:        &e,
		FullName:     &fn,
		Role:         role,
		PasswordHash: passwordHash,
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	mockUserService := setupAuth(t)

	mockUserService.On("Create", mock.AnythingOfType("*model.User")).
		Return(nil).
		Once()

	authSvc := NewAuthService(mockUserService)

	token, err := authSvc.Register(&request.Register{
		Email:    "test@example.com",
		Password: "password123",
		FullName: "Test User",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Register_CreateFails(t *testing.T) {
	mockUserService := setupAuth(t)

	mockUserService.On("Create", mock.AnythingOfType("*model.User")).
		Return(errors.New("db error")).
		Once()

	authSvc := NewAuthService(mockUserService)

	token, err := authSvc.Register(&request.Register{
		Email:    "test@example.com",
		Password: "password123",
		FullName: "Test User",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserService := setupAuth(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := makeUser(1, "test@example.com", "Test User", enum.RoleClient, string(hash))

	mockUserService.On("FindByEmail", "test@example.com").
		Return(user, nil).
		Once()

	authSvc := NewAuthService(mockUserService)

	token, err := authSvc.Login(&request.Login{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserService := setupAuth(t)

	mockUserService.On("FindByEmail", "notfound@example.com").
		Return(nil, errors.New("not found")).
		Once()

	authSvc := NewAuthService(mockUserService)

	token, err := authSvc.Login(&request.Login{
		Email:    "notfound@example.com",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockUserService := setupAuth(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := makeUser(1, "test@example.com", "Test User", enum.RoleClient, string(hash))

	mockUserService.On("FindByEmail", "test@example.com").
		Return(user, nil).
		Once()

	authSvc := NewAuthService(mockUserService)

	token, err := authSvc.Login(&request.Login{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
}
