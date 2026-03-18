package database

import (
	"college-graduation-project-backend/internal/model"

	"gorm.io/gorm"
)

type reviewDatabase struct {
	db *gorm.DB
}

func NewReviewDatabase(database *gorm.DB) ReviewDatabase {
	return &reviewDatabase{db: database}
}

func (d *reviewDatabase) Create(review *model.Review) error {
	return d.db.Create(review).Error
}

func (d *reviewDatabase) FindByID(id uint64) (*model.Review, error) {
	var review model.Review
	if err := d.db.Preload("Booking").Preload("User").First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (d *reviewDatabase) FindAll() ([]model.Review, error) {
	var reviews []model.Review
	if err := d.db.Find(&reviews).Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

func (d *reviewDatabase) Update(review *model.Review) error {
	return d.db.Save(review).Error
}

func (d *reviewDatabase) Delete(id uint64) error {
	return d.db.Delete(&model.Review{}, id).Error
}
