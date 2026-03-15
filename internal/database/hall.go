package database

import (
	"college-graduation-project-backend/internal/model"

	"gorm.io/gorm"
)

type hallDatabase struct {
	db *gorm.DB
}

func NewHallDatabase(database *gorm.DB) HallDatabase {
	return &hallDatabase{db: database}
}

func (d *hallDatabase) Create(hall *model.Hall) error {
	return d.db.Create(hall).Error
}

func (d *hallDatabase) FindByID(id uint64) (*model.Hall, error) {
	var hall model.Hall
	if err := d.db.First(&hall, id).Error; err != nil {
		return nil, err
	}
	return &hall, nil
}

func (d *hallDatabase) FindAll() ([]model.Hall, error) {
	var halls []model.Hall
	if err := d.db.Find(&halls).Error; err != nil {
		return nil, err
	}
	return halls, nil
}

func (d *hallDatabase) Update(hall *model.Hall) error {
	return d.db.Save(hall).Error
}

func (d *hallDatabase) Delete(id uint64) error {
	return d.db.Delete(&model.Hall{}, id).Error
}
