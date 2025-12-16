package yaml

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "github.com/Tutors42Lyon/Mithril/internal/models"
)

type Loader struct {
    parser       *Parser
    exercisesDir string
}

func NewLoader(exercisesDir string) *Loader {
    return &Loader{
        parser:       NewParser(exercisesDir),
        exercisesDir: exercisesDir,
    }
}

// LoadAll scans exercises directory and loads all pools and exercises
func (l *Loader) LoadAll() (map[string]*models.Pool, map[string]*models.Exercise, error) {
    pools := make(map[string]*models.Pool)
    exercises := make(map[string]*models.Exercise)
    errors := []error{}

    // Walk exercises directory
    err := filepath.Walk(l.exercisesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Find pool.yaml files
        if !info.IsDir() && info.Name() == "pool.yaml" {
            pool, err := l.parser.ParsePool(path)
            if err != nil {
                log.Printf("Error parsing pool %s: %v", path, err)
                errors = append(errors, err)
                return nil // Continue walking
            }
            pools[pool.ID] = pool

            // Load exercises in this pool
            poolDir := filepath.Dir(path)
            l.loadPoolExercises(poolDir, pool.ID, exercises, &errors)
        }

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    if len(errors) > 0 {
        return pools, exercises, fmt.Errorf("loaded with %d errors", len(errors))
    }

    return pools, exercises, nil
}

func (l *Loader) loadPoolExercises(poolDir string, poolID string, exercises map[string]*models.Exercise, errors *[]error) {
    // Scan for exercise.yaml files in subdirectories
    filepath.Walk(poolDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && info.Name() == "exercise.yaml" {
            exercise, err := l.parser.ParseExercise(path, poolID)
            if err != nil {
                log.Printf("Error parsing exercise %s: %v", path, err)
                *errors = append(*errors, err)
                return nil
            }
            exercises[exercise.ID] = exercise
        }
        return nil
    })
}

// LoadPool loads a specific pool and its exercises
func (l *Loader) LoadPool(poolID string) (*models.Pool, []*models.Exercise, error) {
	var foundPool *models.Pool
	exercises := []*models.Exercise{}

	// Walk exercises directory to find the pool
	err := filepath.Walk(l.exercisesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "pool.yaml" {
			pool, err := l.parser.ParsePool(path)
			if err != nil {
				return nil // Continue looking
			}

			if pool.ID == poolID {
				foundPool = pool
				// Load exercises in this pool
				poolDir := filepath.Dir(path)
				errors := []error{}
				l.loadPoolExercisesIntoSlice(poolDir, pool.ID, &exercises, &errors)
				return fmt.Errorf("pool found") // Stop walking
			}
		}

		return nil
	})

	if err != nil && err.Error() == "pool found" {
		return foundPool, exercises, nil
	}

	if foundPool == nil {
		return nil, nil, fmt.Errorf("pool not found: %s", poolID)
	}

	return foundPool, exercises, err
}

// loadPoolExercisesIntoSlice is a helper that loads exercises into a slice
func (l *Loader) loadPoolExercisesIntoSlice(poolDir string, poolID string, exercises *[]*models.Exercise, errors *[]error) {
	filepath.Walk(poolDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == "exercise.yaml" {
			exercise, err := l.parser.ParseExercise(path, poolID)
			if err != nil {
				log.Printf("Error parsing exercise %s: %v", path, err)
				*errors = append(*errors, err)
				return nil
			}
			*exercises = append(*exercises, exercise)
		}
		return nil
	})
}

// LoadExercise loads a specific exercise
func (l *Loader) LoadExercise(exerciseID string) (*models.Exercise, error) {
	var foundExercise *models.Exercise

	// Walk exercises directory to find the exercise
	err := filepath.Walk(l.exercisesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "exercise.yaml" {
			// First, find the pool ID by looking for pool.yaml in parent directory
			poolDir := filepath.Dir(filepath.Dir(path))
			poolYamlPath := filepath.Join(poolDir, "pool.yaml")

			pool, err := l.parser.ParsePool(poolYamlPath)
			poolID := ""
			if err == nil {
				poolID = pool.ID
			}

			exercise, err := l.parser.ParseExercise(path, poolID)
			if err != nil {
				return nil // Continue looking
			}

			if exercise.ID == exerciseID {
				foundExercise = exercise
				return fmt.Errorf("exercise found") // Stop walking
			}
		}

		return nil
	})

	if err != nil && err.Error() == "exercise found" {
		return foundExercise, nil
	}

	if foundExercise == nil {
		return nil, fmt.Errorf("exercise not found: %s", exerciseID)
	}

	return foundExercise, err
}

