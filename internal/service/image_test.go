package service

import (
	"college-graduation-project-backend/internal/mocks"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupImage(t *testing.T) (*mocks.MockImageDatabase, *mocks.MockUserDatabase, *mocks.MockHallDatabase) {
	t.Helper()
	return mocks.NewMockImageDatabase(t), mocks.NewMockUserDatabase(t), mocks.NewMockHallDatabase(t)
}

func makeUserDB(id uint64, role enum.UserRole) *model.User {
	e := "user@example.com"
	fn := "User"
	return &model.User{ID: id, Email: &e, FullName: &fn, Role: role}
}

func TestImageService_UploadHallImage_Success(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	admin := makeUserDB(1, enum.RoleAdmin)
	hall := makeHall(1, "Hall A", true)
	image := model.NewImage(0, "/images/hall1.jpg")

	userDB.On("FindByID", uint64(1)).Return(admin, nil)
	hallDB.On("FindByID", uint64(1)).Return(hall, nil)
	imageDB.On("Create", mock.AnythingOfType("*model.Image")).Return(nil)
	imageDB.On("AttachToHall", uint64(1), image.ID).Return(nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.UploadHallImage(1, 1, "/images/hall1.jpg")

	assert.NoError(t, err)
}

func TestImageService_UploadHallImage_NotAdmin(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	client := makeUserDB(2, enum.RoleClient)
	userDB.On("FindByID", uint64(2)).Return(client, nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.UploadHallImage(2, 1, "/images/hall1.jpg")

	assert.Error(t, err)
	imageDB.AssertNotCalled(t, "Create")
}

func TestImageService_UploadHallImage_HallNotFound(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	admin := makeUserDB(1, enum.RoleAdmin)
	userDB.On("FindByID", uint64(1)).Return(admin, nil)
	hallDB.On("FindByID", uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.UploadHallImage(1, 99, "/images/hall1.jpg")

	assert.Error(t, err)
	imageDB.AssertNotCalled(t, "Create")
}

func TestImageService_UploadHallImage_AttachFails_CleansUp(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	admin := makeUserDB(1, enum.RoleAdmin)
	hall := makeHall(1, "Hall A", true)

	userDB.On("FindByID", uint64(1)).Return(admin, nil)
	hallDB.On("FindByID", uint64(1)).Return(hall, nil)
	imageDB.On("Create", mock.AnythingOfType("*model.Image")).Return(nil)
	imageDB.On("AttachToHall", uint64(1), uint64(0)).Return(gorm.ErrInvalidData)
	imageDB.On("Delete", uint64(0)).Return(nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.UploadHallImage(1, 1, "/images/hall1.jpg")

	assert.Error(t, err)
	imageDB.AssertCalled(t, "Delete", uint64(0))
}

func TestImageService_GetHallImages_Success(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	hall := makeHall(1, "Hall A", true)
	images := []model.Image{
		{ID: 1, Path: "/images/hall1.jpg"},
		{ID: 2, Path: "/images/hall2.jpg"},
	}

	hallDB.On("FindByID", uint64(1)).Return(hall, nil)
	imageDB.On("FindByHallID", uint64(1)).Return(images, nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	result, err := svc.GetHallImages(1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "/images/hall1.jpg", result[0])
}

func TestImageService_GetHallImages_HallNotFound(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	hallDB.On("FindByID", uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewImageService(imageDB, userDB, hallDB)

	result, err := svc.GetHallImages(99)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestImageService_SetUserAvatar_Success_NoOldAvatar(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	user := makeUserDB(1, enum.RoleClient)

	userDB.On("FindByID", uint64(1)).Return(user, nil)
	imageDB.On("GetByUserID", uint64(1)).Return(nil, gorm.ErrRecordNotFound)
	imageDB.On("Create", mock.AnythingOfType("*model.Image")).Return(nil)
	imageDB.On("SetUserImage", uint64(1), uint64(0)).Return(nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.SetUserAvatar(1, "/avatars/user1.jpg")

	assert.NoError(t, err)
}

func TestImageService_SetUserAvatar_ReplacesOldAvatar(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	user := makeUserDB(1, enum.RoleClient)
	oldImage := &model.Image{ID: 10, Path: "/avatars/old.jpg"}

	userDB.On("FindByID", uint64(1)).Return(user, nil)
	imageDB.On("GetByUserID", uint64(1)).Return(oldImage, nil)
	imageDB.On("Create", mock.AnythingOfType("*model.Image")).Return(nil)
	imageDB.On("SetUserImage", uint64(1), uint64(0)).Return(nil)
	imageDB.On("Delete", uint64(10)).Return(nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.SetUserAvatar(1, "/avatars/new.jpg")

	assert.NoError(t, err)
	imageDB.AssertCalled(t, "Delete", uint64(10))
}

func TestImageService_GetUserAvatar_Success(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	user := makeUserDB(1, enum.RoleClient)
	image := &model.Image{ID: 1, Path: "/avatars/user1.jpg"}

	userDB.On("FindByID", uint64(1)).Return(user, nil)
	imageDB.On("GetByUserID", uint64(1)).Return(image, nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	result, err := svc.GetUserAvatar(1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "/avatars/user1.jpg", *result)
}

func TestImageService_GetUserAvatar_NoAvatar(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	user := makeUserDB(1, enum.RoleClient)

	userDB.On("FindByID", uint64(1)).Return(user, nil)
	imageDB.On("GetByUserID", uint64(1)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewImageService(imageDB, userDB, hallDB)

	result, err := svc.GetUserAvatar(1)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestImageService_DeleteUserAvatar_Success(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	user := makeUserDB(1, enum.RoleClient)
	image := &model.Image{ID: 5, Path: "/avatars/user1.jpg"}

	userDB.On("FindByID", uint64(1)).Return(user, nil)
	imageDB.On("GetByUserID", uint64(1)).Return(image, nil)
	imageDB.On("RemoveUserImage", uint64(1)).Return(nil)
	imageDB.On("Delete", uint64(5)).Return(nil)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.DeleteUserAvatar(1)

	assert.NoError(t, err)
}

func TestImageService_DeleteUserAvatar_NoAvatar(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	user := makeUserDB(1, enum.RoleClient)

	userDB.On("FindByID", uint64(1)).Return(user, nil)
	imageDB.On("GetByUserID", uint64(1)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.DeleteUserAvatar(1)

	assert.NoError(t, err)
	imageDB.AssertNotCalled(t, "RemoveUserImage")
}

func TestImageService_UserNotFound(t *testing.T) {
	imageDB, userDB, hallDB := setupImage(t)

	userDB.On("FindByID", uint64(99)).Return(nil, gorm.ErrRecordNotFound)

	svc := NewImageService(imageDB, userDB, hallDB)

	err := svc.SetUserAvatar(99, "/avatars/user.jpg")

	assert.Error(t, err)
	imageDB.AssertNotCalled(t, "Create")
}
