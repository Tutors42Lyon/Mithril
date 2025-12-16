package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/Tutors42Lyon/Mithril/internal/cache"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"github.com/Tutors42Lyon/Mithril/internal/yaml"
	"github.com/nats-io/nats.go"
)

type ExerciseService struct {
	cache    *cache.ExerciseCache
	loader   *yaml.Loader
	natsConn *nats.Conn
}

func NewExerciseService(cache *cache.ExerciseCache, loader *yaml.Loader, nc *nats.Conn) *ExerciseService {
	return &ExerciseService{
		cache:    cache,
		loader:   loader,
		natsConn: nc,
	}
}

// GetPoolList retrieves all pools (cache-first)
func (es *ExerciseService) GetPoolList() ([]*models.Pool, error) {
	poolIDs, err := es.cache.ListPools()
	if err != nil || len(poolIDs) == 0 {
		// Cache miss - reload from YAML
		pools, _, err := es.loader.LoadAll()
		return poolsMapToSlice(pools), err
	}

	// Get each pool from cache
	pools := []*models.Pool{}
	for _, poolID := range poolIDs {
		pool, err := es.cache.GetPool(poolID)
		if err != nil {
			continue
		}
		pools = append(pools, pool)
	}
	return pools, nil
}

// GetPool retrieves a specific pool
func (es *ExerciseService) GetPool(poolID string) (*models.Pool, error) {
	pool, err := es.cache.GetPool(poolID)
	if err != nil {
		// Cache miss - load from YAML
		pool, _, err = es.loader.LoadPool(poolID)
		if err != nil {
			return nil, err
		}
		// Cache for future
		es.cache.SetPool(pool, 1*time.Hour)
	}
	return pool, nil
}

// GetExercise retrieves a specific exercise
func (es *ExerciseService) GetExercise(exerciseID string) (*models.Exercise, error) {
	exercise, err := es.cache.GetExercise(exerciseID)
	if err != nil {
		// Cache miss - load from YAML
		exercise, err = es.loader.LoadExercise(exerciseID)
		if err != nil {
			return nil, err
		}
		es.cache.SetExercise(exercise, 1*time.Hour)
	}
	return exercise, nil
}

// GetPoolExercises retrieves all exercises in a pool
func (es *ExerciseService) GetPoolExercises(poolID string) ([]*models.Exercise, error) {
	exerciseIDs, err := es.cache.GetPoolExercises(poolID)
	if err != nil || len(exerciseIDs) == 0 {
		// Cache miss - load from YAML
		_, exercises, err := es.loader.LoadPool(poolID)
		return exercises, err
	}

	// Get each exercise from cache
	exercises := []*models.Exercise{}
	for _, exerciseID := range exerciseIDs {
		exercise, err := es.cache.GetExercise(exerciseID)
		if err != nil {
			continue
		}
		exercises = append(exercises, exercise)
	}
	return exercises, nil
}

// LaunchExercise creates a session for user to start an exercise
func (es *ExerciseService) LaunchExercise(userID string, exerciseID string) (*models.Exercise, string, error) {
	exercise, err := es.GetExercise(exerciseID)
	if err != nil {
		return nil, "", err
	}

	sessionID := generateSessionID()

	// Publish exercise.{userID}.status with "started" status
	es.publishStatus(userID, exerciseID, "started")

	return exercise, sessionID, nil
}

// ReloadExercise reloads exercise from YAML and updates cache
func (es *ExerciseService) ReloadExercise(poolID, exerciseID string) error {
	// Load exercise from YAML
	exercise, err := es.loader.LoadExercise(exerciseID)
	if err != nil {
		return err
	}

	// Update cache
	if err := es.cache.SetExercise(exercise, 1*time.Hour); err != nil {
		return err
	}

	// Publish update notification
	updateMsg := map[string]string{
		"exercise_id": exerciseID,
		"pool_id":     poolID,
	}
	updateJSON, _ := json.Marshal(updateMsg)
	es.natsConn.Publish("exercise.updated", updateJSON)

	return nil
}

func (es *ExerciseService) publishStatus(userID, exerciseID, status string) {
	// Publish to NATS topic for WebSocket forwarding
	statusMsg := map[string]interface{}{
		"user_id":     userID,
		"exercise_id": exerciseID,
		"status":      status,
		"timestamp":   time.Now(),
	}
	statusJSON, _ := json.Marshal(statusMsg)
	es.natsConn.Publish("exercise."+userID+".status", statusJSON)
}

// Helper function to generate unique session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Helper function to convert pool map to slice
func poolsMapToSlice(poolsMap map[string]*models.Pool) []*models.Pool {
	pools := make([]*models.Pool, 0, len(poolsMap))
	for _, pool := range poolsMap {
		pools = append(pools, pool)
	}
	return pools
}

