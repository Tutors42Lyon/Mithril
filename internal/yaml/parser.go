package yaml

import (
	"fmt"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"gopkg.in/yaml.v2"
	"os"
)

type Parser struct {
	exercisesDir string
}

func NewParser(exercisesDir string) *Parser {
	return &Parser{exercisesDir: exercisesDir}
}

func (p *Parser) ParsePool(poolPath string) (*models.Pool, error) {
	data, err := os.ReadFile(poolPath)
	if err != nil {
		return nil, err
	}

	var pool models.Pool
	if err := yaml.Unmarshal(data, &pool); err != nil {
		return nil, fmt.Errorf("failed to parse pool %s: %w", poolPath, err)
	}

	return &pool, nil
}

func (p *Parser) ParseExercise(exercisePath string, poolID string) (*models.Exercise, error) {
	data, err := os.ReadFile(exercisePath)
	if err != nil {
		return nil, err
	}

	var exercise models.Exercise
	if err := yaml.Unmarshal(data, &exercise); err != nil {
		return nil, fmt.Errorf("failed to parse exercise %s: %w", exercisePath, err)
	}

	exercise.PoolID = poolID
	return &exercise, nil
}

