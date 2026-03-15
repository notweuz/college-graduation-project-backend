package database

import (
	"college-graduation-project-backend/internal/model"

	"gorm.io/gorm"
)

type userDatabase struct {
	db *gorm.DB
}

func NewUserDatabase(database *gorm.DB) UserDatabase {
	return &userDatabase{db: database}
}

func (d *userDatabase) Create(user *model.User) error {
	return d.db.Create(user).Error
}

func (d *userDatabase) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := d.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *userDatabase) FindAll() ([]model.User, error) {
	var users []model.User
	if err := d.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (d *userDatabase) Update(user *model.User) error {
	return d.db.Save(user).Error
}

func (d *userDatabase) Delete(id uint) error {
	return d.db.Delete(&model.User{}, id).Error
}
