package repository

import (
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"gorm.io/gorm"

)

type PoolRepository struct {
	DB *gorm.DB
}

func NewPoolRepository(db *gorm.DB) *PoolRepository {
	return &PoolRepository{DB: db}
}

func (r *PoolRepository) GetAll() (*[]models.PoolMessage, error) {
    var pools []models.PoolMessage

	result := r.DB.Find(&pools)
	if result.Error != nil {
		return nil, result.Error
	}
	return &pools, nil
}
