package repository

import (
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"gorm.io/gorm"

)

type ExerciseRepository struct {
	DB *gorm.DB
}

func NewExerciseRepository(db *gorm.DB) *ExerciseRepository {
	return &ExerciseRepository{DB: db}
}

func (r *ExerciseRepository) GetExercises(id int) (*[]models.ExerciseMessage, error) {
	var exercises []models.ExerciseMessage

	result := r.DB.Where("id = ?", id).Find(&exercises)

	return &exercises, result.Error

}
