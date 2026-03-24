package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type userAgreementService struct {
	database    database.UserAgreementDatabase
	userService UserService
}

func NewUserAgreementService(database database.UserAgreementDatabase, userService UserService) UserAgreementService {
	return &userAgreementService{database: database, userService: userService}
}

func (s *userAgreementService) Get() (*model.UserAgreement, error) {
	agreement, err := s.database.Get()
	if err != nil {
		log.Error().Err(err).Msg("Cannot get user agreement")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot get user agreement", "user agreement is not set")
		}
		return nil, errs.InternalServerError("Cannot get user agreement", "internal server error")
	}

	return agreement, nil
}

func (s *userAgreementService) Update(userID uint64, text string) (*model.UserAgreement, error) {
	user, err := s.userService.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Role != enum.RoleAdmin {
		log.Warn().Uint64("id", userID).Msg("User is not admin")
		return nil, errs.Forbidden("Forbidden", "user is not admin")
	}

	agreement, err := s.database.Save(text)
	if err != nil {
		log.Error().Err(err).Msg("Cannot update user agreement")
		return nil, errs.InternalServerError("Cannot update user agreement", "internal server error")
	}

	return agreement, nil
}
