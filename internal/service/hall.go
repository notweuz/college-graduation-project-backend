package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type hallService struct {
	database database.HallDatabase
}

func NewHallService(database database.HallDatabase) HallService {
	return &hallService{database: database}
}

func (s *hallService) Create(hall *model.Hall) error {
	err := s.database.Create(hall)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errs.Conflict("Cannot create hall", "hall already exists")
		}
		return errs.InternalServerError("Cannot create hall", "internal server error")
	}
	return nil
}

func (s *hallService) FindByID(id uint64) (*model.Hall, error) {
	hall, err := s.database.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot find hall", "hall with that id doesnt exist")
		}
		return nil, errs.InternalServerError("Cannot find hall", "internal server error")
	}
	return hall, nil
}

func (s *hallService) FindAll() ([]model.Hall, error) {
	halls, err := s.database.FindAll()
	if err != nil {
		return nil, errs.InternalServerError("Cannot get halls", "internal server error")
	}
	return halls, nil
}

func (s *hallService) Update(hall *model.Hall) error {
	err := s.database.Update(hall)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot update hall", "hall with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot update hall", "internal server error")
	}
	return nil
}

func (s *hallService) Delete(id uint64) error {
	err := s.database.Delete(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete hall", "hall with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot delete hall", "internal server error")
	}
	return nil
}
