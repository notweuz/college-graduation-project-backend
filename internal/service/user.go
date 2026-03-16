package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/request"
	"errors"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type userService struct {
	database database.UserDatabase
}

func NewUserService(database database.UserDatabase) UserService {
	return &userService{database: database}
}

func (u *userService) Create(user *model.User) error {
	if _, err := u.FindByEmail(*user.Email); err == nil {
		return errs.Conflict("Cannot create user", "user with this email already exists")
	}

	err := u.database.Create(user)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.Conflict("Cannot create user", "user already exists")
		}
		return errs.InternalServerError("Cannot create user", "internal server error")
	}
	return nil
}

func (u *userService) FindByID(id uint64) (*model.User, error) {
	user, err := u.database.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot find user", "user with that id doesnt exist in database")
		}
		return nil, errs.InternalServerError("Cannot get user", "internal server error")
	}
	return user, nil
}

func (u *userService) FindAll() ([]model.User, error) {
	users, err := u.database.FindAll()
	if err != nil {
		return nil, errs.InternalServerError("Cannot get users", "internal server error")
	}
	return users, nil
}

func (u *userService) Update(user *model.User) error {
	err := u.database.Update(user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot update user", "user with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot update user", "internal server error")
	}
	return nil
}

func (u *userService) Delete(id uint64) error {
	err := u.database.Delete(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete user", "user with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot delete user", "internal server error")
	}
	return nil
}

func (u *userService) FindByEmail(email string) (*model.User, error) {
	user, err := u.database.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userService) UpdateProfile(ctx fiber.Ctx, updateProfile request.UpdateProfile) (*model.User, error) {
	userId, err := middleware.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	user, err := u.FindByID(userId)
	if err != nil {
		return nil, err
	}

	if updateProfile.FullName != nil {
		user.FullName = updateProfile.FullName
	}
	if updateProfile.Email != nil {
		user.Email = updateProfile.Email
	}

	err = u.Update(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
