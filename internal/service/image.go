package service

import (
	"college-graduation-project-backend/internal/database"
	"college-graduation-project-backend/internal/errs"
	"college-graduation-project-backend/internal/model"
	"college-graduation-project-backend/internal/model/enum"
	"college-graduation-project-backend/internal/model/response"
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type imageService struct {
	imageDatabase database.ImageDatabase
	userDatabase  database.UserDatabase
	hallDatabase  database.HallDatabase
}

func NewImageService(imageDatabase database.ImageDatabase, userDatabase database.UserDatabase, hallDatabase database.HallDatabase) ImageService {
	return &imageService{
		imageDatabase: imageDatabase,
		userDatabase:  userDatabase,
		hallDatabase:  hallDatabase,
	}
}

func (s *imageService) UploadHallImage(userID, hallID uint64, imagePath string) error {
	user, err := s.userDatabase.FindByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot find user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot upload hall image", "user with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot upload hall image", "internal server error")
	}

	if user.Role != enum.RoleAdmin {
		log.Warn().Uint64("id", userID).Msg("User is not admin")
		return errs.Forbidden("Forbidden", "user is not admin")
	}

	_, err = s.hallDatabase.FindByID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot find hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot upload hall image", "hall with that id doesnt exist")
		}
		return errs.InternalServerError("Cannot upload hall image", "internal server error")
	}

	hallImage := model.NewImage(0, imagePath)
	err = s.imageDatabase.Create(hallImage)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot create image record")
		return errs.InternalServerError("Cannot upload hall image", "internal server error")
	}

	err = s.imageDatabase.AttachToHall(hallID, hallImage.ID)
	if err != nil {
		_ = s.imageDatabase.Delete(hallImage.ID)
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot link hall image")
		return errs.InternalServerError("Cannot upload hall image", "internal server error")
	}

	log.Info().Uint64("hallID", hallID).Str("path", imagePath).Msg("Hall image successfully uploaded")
	return nil
}

func (s *imageService) DeleteHallImage(userID, hallID, imageID uint64) error {
	user, err := s.userDatabase.FindByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot find user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete hall image", "user with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot delete hall image", "internal server error")
	}

	if user.Role != enum.RoleAdmin {
		log.Warn().Uint64("id", userID).Msg("User is not admin")
		return errs.Forbidden("Forbidden", "user is not admin")
	}

	_, err = s.hallDatabase.FindByID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot find hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete hall image", "hall with that id doesnt exist")
		}
		return errs.InternalServerError("Cannot delete hall image", "internal server error")
	}

	image, err := s.imageDatabase.GetByID(imageID)
	if err != nil {
		log.Error().Err(err).Uint64("imageID", imageID).Msg("Cannot find image")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete hall image", "image with that id doesnt exist")
		}
		return errs.InternalServerError("Cannot delete hall image", "internal server error")
	}

	hallImages, err := s.imageDatabase.FindByHallID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot get hall images")
		return errs.InternalServerError("Cannot delete hall image", "internal server error")
	}

	found := false
	for _, img := range hallImages {
		if img.ID == imageID {
			found = true
			break
		}
	}
	if !found {
		return errs.NotFound("Cannot delete hall image", "image is not attached to this hall")
	}

	err = s.imageDatabase.Delete(imageID)
	if err != nil {
		log.Error().Err(err).Uint64("imageID", imageID).Msg("Cannot delete image")
		return errs.InternalServerError("Cannot delete hall image", "internal server error")
	}

	if image.Path != "" {
		filePath := "." + image.Path
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Warn().Err(err).Str("path", filePath).Msg("Cannot delete image file")
		}
	}

	log.Info().Uint64("hallID", hallID).Uint64("imageID", imageID).Msg("Hall image successfully deleted")
	return nil
}

func (s *imageService) GetHallImages(hallID uint64) ([]string, error) {
	_, err := s.hallDatabase.FindByID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot find hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot get hall images", "hall with that id doesnt exist")
		}
		return nil, errs.InternalServerError("Cannot get hall images", "internal server error")
	}

	images, err := s.imageDatabase.FindByHallID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot get hall images")
		return nil, errs.InternalServerError("Cannot get hall images", "internal server error")
	}

	imagePaths := make([]string, 0, len(images))
	for _, image := range images {
		imagePaths = append(imagePaths, image.Path)
	}

	log.Info().Uint64("hallID", hallID).Int("count", len(imagePaths)).Msg("Hall images successfully retrieved")
	return imagePaths, nil
}

func (s *imageService) GetHallImagesWithIDs(hallID uint64) ([]response.HallImageWithID, error) {
	_, err := s.hallDatabase.FindByID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot find hall")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot get hall images", "hall with that id doesnt exist")
		}
		return nil, errs.InternalServerError("Cannot get hall images", "internal server error")
	}

	images, err := s.imageDatabase.FindByHallID(hallID)
	if err != nil {
		log.Error().Err(err).Uint64("hallID", hallID).Msg("Cannot get hall images")
		return nil, errs.InternalServerError("Cannot get hall images", "internal server error")
	}

	result := make([]response.HallImageWithID, 0, len(images))
	for _, image := range images {
		result = append(result, response.HallImageWithID{ID: image.ID, Path: image.Path})
	}

	log.Info().Uint64("hallID", hallID).Int("count", len(result)).Msg("Hall images with IDs successfully retrieved")
	return result, nil
}

func (s *imageService) SetUserAvatar(userID uint64, imagePath string) error {
	_, err := s.userDatabase.FindByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot find user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot set user avatar", "user with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot set user avatar", "internal server error")
	}

	var oldImageID uint64
	oldImage, err := s.imageDatabase.GetByUserID(userID)
	if err == nil && oldImage != nil {
		oldImageID = oldImage.ID
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot get old user avatar")
		return errs.InternalServerError("Cannot set user avatar", "internal server error")
	}

	newImage := model.NewImage(0, imagePath)
	if err := s.imageDatabase.Create(newImage); err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot create image record")
		return errs.InternalServerError("Cannot set user avatar", "internal server error")
	}

	if err := s.imageDatabase.SetUserImage(userID, newImage.ID); err != nil {
		_ = s.imageDatabase.Delete(newImage.ID)
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot set user image link")
		return errs.InternalServerError("Cannot set user avatar", "internal server error")
	}

	if oldImageID != 0 {
		if err := s.imageDatabase.Delete(oldImageID); err != nil {
			log.Warn().Err(err).Uint64("id", userID).Uint64("imageID", oldImageID).Msg("Cannot delete old avatar image")
		}
	}

	log.Info().Uint64("id", userID).Str("path", imagePath).Msg("User avatar successfully updated")
	return nil
}

func (s *imageService) GetUserAvatar(userID uint64) (*string, error) {
	_, err := s.userDatabase.FindByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot find user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Cannot get user avatar", "user with that id doesnt exist in database")
		}
		return nil, errs.InternalServerError("Cannot get user avatar", "internal server error")
	}

	image, err := s.imageDatabase.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot get user avatar")
		return nil, errs.InternalServerError("Cannot get user avatar", "internal server error")
	}

	return &image.Path, nil
}

func (s *imageService) DeleteUserAvatar(userID uint64) error {
	_, err := s.userDatabase.FindByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot find user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("Cannot delete user avatar", "user with that id doesnt exist in database")
		}
		return errs.InternalServerError("Cannot delete user avatar", "internal server error")
	}

	image, err := s.imageDatabase.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot get user avatar")
		return errs.InternalServerError("Cannot delete user avatar", "internal server error")
	}

	if err := s.imageDatabase.RemoveUserImage(userID); err != nil {
		log.Error().Err(err).Uint64("id", userID).Msg("Cannot remove user avatar link")
		return errs.InternalServerError("Cannot delete user avatar", "internal server error")
	}

	if err := s.imageDatabase.Delete(image.ID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Warn().Err(err).Uint64("id", userID).Uint64("imageID", image.ID).Msg("Cannot delete avatar image record")
	}

	log.Info().Uint64("id", userID).Msg("User avatar successfully deleted")
	return nil
}
