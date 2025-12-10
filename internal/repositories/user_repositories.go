package repository

import (
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user *models.UserMessage) error {
	result := r.DB.Create(user)
	return result.Error
}