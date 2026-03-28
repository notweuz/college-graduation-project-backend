package database

import (
	"college-graduation-project-backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type imageDatabase struct {
	db *gorm.DB
}

func NewImageDatabase(db *gorm.DB) ImageDatabase {
	return &imageDatabase{db: db}
}

func (h *imageDatabase) Create(image *model.Image) error {
	return h.db.Create(image).Error
}

func (h *imageDatabase) GetByID(id uint64) (*model.Image, error) {
	var image model.Image
	err := h.db.First(&image, id).Error
	return &image, err
}

func (h *imageDatabase) FindByHallID(hallID uint64) ([]model.Image, error) {
	var images []model.Image
	err := h.db.
		Model(&model.Image{}).
		Joins("JOIN hall_images ON hall_images.image_id = images.id").
		Where("hall_images.hall_id = ?", hallID).
		Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (h *imageDatabase) GetByUserID(userID uint64) (*model.Image, error) {
	var userImage model.UserImage
	err := h.db.
		Preload("Image").
		Where("user_id = ?", userID).
		First(&userImage).Error
	if err != nil {
		return nil, err
	}
	return &userImage.Image, nil
}

func (h *imageDatabase) AttachToHall(hallID, imageID uint64) error {
	return h.db.Create(model.NewHallImage(hallID, imageID)).Error
}

func (h *imageDatabase) SetUserImage(userID, imageID uint64) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserImage{}).Error; err != nil {
			return err
		}
		return tx.Create(model.NewUserImage(userID, imageID)).Error
	})
}

func (h *imageDatabase) RemoveUserImage(userID uint64) error {
	return h.db.Where("user_id = ?", userID).Delete(&model.UserImage{}).Error
}

func (h *imageDatabase) Delete(id uint64) error {
	txErr := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("image_id = ?", id).Delete(&model.HallImage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("image_id = ?", id).Delete(&model.UserImage{}).Error; err != nil {
			return err
		}

		res := tx.Delete(&model.Image{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})

	if txErr != nil && errors.Is(txErr, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}

	return txErr
}
