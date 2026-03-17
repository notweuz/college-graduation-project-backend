package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/middleware"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/request"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type hallService struct {
	database    database.HallDatabase
	userService UserService
}

func NewHallService(database database.HallDatabase, userService UserService) HallService {
	return &hallService{database: database, userService: userService}
}

func (s *hallService) Create(ctx fiber.Ctx, hallCreate *request.HallCreate) (*model.Hall, error) {
	userId, err := middleware.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.userService.FindByID(userId)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleAdmin {
		log.Error().Uint64("id", userId).Msg("User is not admin")
		return nil, errs.Forbidden("Forbidden", "user is not admin")
	}

	hall := model.NewHall(0, hallCreate.Name, hallCreate.Description, hallCreate.PricePerHour, true)

	err = s.database.Create(hall)
	if err != nil {
		log.Error().Err(err).Msg("Cannot create hall")
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errs.Conflict("Cannot create hall", "hall already exists")
		}
		return nil, errs.InternalServerError("Cannot create hall", "internal server error")
	}
	log.Info().Uint64("id", hall.ID).Str("name", hall.Name).Msg("Hall successfully created")
	return hall, nil
}

func (s *hallService) FindByID(id uint64) (*model.Hall, error) {
	hall, err := s.database.FindByID(id)
	if err != nil {
		log.Error().Err(err).Msg("Cannot find hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot find hall", "hall with that id doesnt exist")
		}
		return nil, errs.InternalServerError("Cannot find hall", "internal server error")
	}
	log.Info().Uint64("id", hall.ID).Str("name", hall.Name).Msg("Hall successfully found")
	return hall, nil
}

func (s *hallService) FindAll() ([]model.Hall, error) {
	halls, err := s.database.FindAll()
	if err != nil {
		log.Error().Err(err).Msg("Cannot get halls")
		return nil, errs.InternalServerError("Cannot get halls", "internal server error")
	}
	log.Info().Int("amount", len(halls)).Msg("Halls successfully found")
	return halls, nil
}

func (s *hallService) FindAllActive() ([]model.Hall, error) {
	halls, err := s.database.FindAllActive()
	if err != nil {
		log.Error().Err(err).Msg("Cannot get active halls")
		return nil, errs.InternalServerError("Cannot get active halls", "internal server error")
	}
	log.Info().Int("amount", len(halls)).Msg("Active halls successfully found")
	return halls, nil
}

func (s *hallService) Update(ctx fiber.Ctx, hallUpdate *request.HallUpdate) (*model.Hall, error) {
	id := fiber.Params[uint64](ctx, "id")
	userId, err := middleware.GetCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}
	user, err := s.userService.FindByID(userId)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleAdmin {
		log.Error().Uint64("id", userId).Msg("User is not admin")
		return nil, errs.Forbidden("Forbidden", "user is not admin")
	}
	hall, err := s.FindByID(id)
	if err != nil {
		return nil, err
	}
	if hallUpdate.Name != nil {
		hall.Name = *hallUpdate.Name
	}
	if hallUpdate.Description != nil {
		hall.Description = *hallUpdate.Description
	}
	if hallUpdate.PricePerHour != nil {
		hall.PricePerHour = *hallUpdate.PricePerHour
	}
	if hallUpdate.IsActive != nil {
		hall.IsActive = *hallUpdate.IsActive
	}
	err = s.database.Update(hall)
	if err != nil {
		log.Error().Err(err).Msg("Cannot update hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot update hall", "hall with that id doesnt exist in database")
		}
		return nil, errs.InternalServerError("Cannot update hall", "internal server error")
	}
	log.Info().Uint64("id", hall.ID).Str("name", hall.Name).Msg("Hall successfully updated")
	return hall, nil
}

func (s *hallService) Delete(id uint64) error {
	err := s.database.Delete(id)
	if err != nil {
		log.Error().Err(err).Msg("Cannot delete hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete hall", "hall with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot delete hall", "internal server error")
	}
	log.Info().Uint64("id", id).Msg("Hall successfully deleted")
	return nil
}
