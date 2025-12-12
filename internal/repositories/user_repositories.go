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
    result := r.DB.Where(models.UserMessage{IntraID: user.IntraID}).
		FirstOrCreate(user)

    return result.Error
}

func (r *UserRepository) UpdateUserRoleByUsername(username string, newRole string) error {
	result := r.DB.Model(&models.UserMessage{}).
		Where("username = ?", username).
		Update("role", newRole)

	return result.Error
}

func (r *UserRepository) GetByUsername(username string) (*models.UserMessage, error) {
    var user models.UserMessage
    result := r.DB.Where("username = ?", username).
		First(&user)
    if result.Error != nil {
        return nil, result.Error
    }
    return &user, nil
}