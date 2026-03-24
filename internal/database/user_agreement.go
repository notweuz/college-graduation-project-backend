package database

import (
	"college-graduation-project-backend/internal/model"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userAgreementDatabase struct {
	db *gorm.DB
}

func NewUserAgreementDatabase(database *gorm.DB) UserAgreementDatabase {
	return &userAgreementDatabase{db: database}
}

func (d *userAgreementDatabase) Get() (*model.UserAgreement, error) {
	var agreement model.UserAgreement
	if err := d.db.Order("id DESC").First(&agreement).Error; err != nil {
		return nil, err
	}
	return &agreement, nil
}

func (d *userAgreementDatabase) Save(text string) (*model.UserAgreement, error) {
	var agreement model.UserAgreement

	err := d.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id DESC").First(&agreement).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				agreement = model.UserAgreement{Text: text, Version: 1}
				return tx.Create(&agreement).Error
			}
			return err
		}

		agreement.Text = text
		agreement.Version++
		return tx.Save(&agreement).Error
	})
	if err != nil {
		return nil, err
	}

	return &agreement, nil
}
