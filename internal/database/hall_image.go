package database

import (
	"college-graduation-project-backend/internal/model"
	"gorm.io/gorm"
)

type hallImageDatabase struct {
	db *gorm.DB
}

func NewHallImageDatabase(db *gorm.DB) HallImageDatabase {
	return &hallImageDatabase{db: db}
}

func (h *hallImageDatabase) Create(hallImage *model.HallImage) error {
	return h.db.Create(hallImage).Error
}

func (h *hallImageDatabase) FindByHallID(hallID uint64) ([]model.HallImage, error) {
	var images []model.HallImage
	err := h.db.Where("hall_id = ?", hallID).Find(&images).Error
	return images, err
}

func (h *hallImageDatabase) Delete(id uint64) error {
	return h.db.Delete(&model.HallImage{}, id).Error
}
