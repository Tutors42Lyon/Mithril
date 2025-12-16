package cache

import "fmt"

const (
	KeyPrefixPool     = "mithril:pool:"
	KeyPrefixExercise = "mithril:exercise:"
	KeyPoolsList      = "mithril:pools:list"
	KeyPoolExercises  = "mithril:pool:%s:exercises" // Format with pool_id
)

func PoolKey(poolID string) string {
	return KeyPrefixPool + poolID
}

func ExerciseKey(exerciseID string) string {
	return KeyPrefixExercise + exerciseID
}

func PoolExercisesKey(poolID string) string {
	return fmt.Sprintf(KeyPoolExercises, poolID)
}

