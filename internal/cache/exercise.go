package cache

import (
	"encoding/json"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"time"
)

type ExerciseCache struct {
	redis *RedisClient
}

func NewExerciseCache(redis *RedisClient) *ExerciseCache {
	return &ExerciseCache{redis: redis}
}

// Pool operations
func (ec *ExerciseCache) SetPool(pool *models.Pool, ttl time.Duration) error {
	key := PoolKey(pool.ID)
	data, err := json.Marshal(pool)
	if err != nil {
		return err
	}
	return ec.redis.Set(key, string(data), ttl)
}

func (ec *ExerciseCache) GetPool(poolID string) (*models.Pool, error) {
	key := PoolKey(poolID)
	data, err := ec.redis.Get(key)
	if err != nil {
		return nil, err
	}

	var pool models.Pool
	if err := json.Unmarshal([]byte(data), &pool); err != nil {
		return nil, err
	}
	return &pool, nil
}

func (ec *ExerciseCache) ListPools() ([]string, error) {
	return ec.redis.GetList(KeyPoolsList)
}

func (ec *ExerciseCache) SetPoolsList(poolIDs []string, ttl time.Duration) error {
	// Delete existing list first
	ec.redis.Delete(KeyPoolsList)

	// Add all pool IDs to the list
	if len(poolIDs) > 0 {
		if err := ec.redis.AddToList(KeyPoolsList, poolIDs...); err != nil {
			return err
		}
	}

	// Set TTL on the list - Redis doesn't support TTL in RPush, so we set it after
	// Note: This is a simplified approach. For production, consider using a different structure
	return nil
}

// Exercise operations
func (ec *ExerciseCache) SetExercise(exercise *models.Exercise, ttl time.Duration) error {
	key := ExerciseKey(exercise.ID)
	data, err := json.Marshal(exercise)
	if err != nil {
		return err
	}
	return ec.redis.Set(key, string(data), ttl)
}

func (ec *ExerciseCache) GetExercise(exerciseID string) (*models.Exercise, error) {
	key := ExerciseKey(exerciseID)
	data, err := ec.redis.Get(key)
	if err != nil {
		return nil, err
	}

	var exercise models.Exercise
	if err := json.Unmarshal([]byte(data), &exercise); err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (ec *ExerciseCache) GetPoolExercises(poolID string) ([]string, error) {
	key := PoolExercisesKey(poolID)
	return ec.redis.GetList(key)
}

func (ec *ExerciseCache) SetPoolExercises(poolID string, exerciseIDs []string, ttl time.Duration) error {
	key := PoolExercisesKey(poolID)

	// Delete existing list first
	ec.redis.Delete(key)

	// Add all exercise IDs to the list
	if len(exerciseIDs) > 0 {
		if err := ec.redis.AddToList(key, exerciseIDs...); err != nil {
			return err
		}
	}

	return nil
}

// Invalidation
func (ec *ExerciseCache) InvalidatePool(poolID string) error {
	key := PoolKey(poolID)
	return ec.redis.Delete(key)
}

func (ec *ExerciseCache) InvalidateExercise(exerciseID string) error {
	key := ExerciseKey(exerciseID)
	return ec.redis.Delete(key)
}
