package messaging

import (
	"encoding/json"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	"github.com/Tutors42Lyon/Mithril/internal/services"
	"github.com/nats-io/nats.go"
)

func LoadExerciseSubscribers(nc *nats.Conn, exerciseService *services.ExerciseService) {
	// Use queue groups for load balancing (like worker service)
	nc.QueueSubscribe("exercise.pool.list", "exercise.service", HandlePoolList(exerciseService))
	nc.QueueSubscribe("exercise.pool.get", "exercise.service", HandlePoolGet(exerciseService))
	nc.QueueSubscribe("exercise.pool.exercises", "exercise.service", HandlePoolExercises(exerciseService))
	nc.QueueSubscribe("exercise.get", "exercise.service", HandleExerciseGet(exerciseService))
	nc.QueueSubscribe("exercise.launch", "exercise.service", HandleExerciseLaunch(exerciseService))
}

type PoolListResponse struct {
	Pools []*models.Pool `json:"pools"`
}

func HandlePoolList(es *services.ExerciseService) nats.MsgHandler {
	return func(m *nats.Msg) {
		pools, err := es.GetPoolList()
		if err != nil {
			m.Respond([]byte(`{"error": "Failed to get pool list"}`))
			return
		}

		response := PoolListResponse{Pools: pools}
		responseJSON, _ := json.Marshal(response)
		m.Respond(responseJSON)
	}
}

type PoolGetRequest struct {
	PoolID string `json:"pool_id"`
}

type PoolGetResponse struct {
	Pool *models.Pool `json:"pool"`
}

func HandlePoolGet(es *services.ExerciseService) nats.MsgHandler {
	return func(m *nats.Msg) {
		var req PoolGetRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			m.Respond([]byte(`{"error": "Invalid request"}`))
			return
		}

		pool, err := es.GetPool(req.PoolID)
		if err != nil {
			m.Respond([]byte(`{"error": "Pool not found"}`))
			return
		}

		response := PoolGetResponse{Pool: pool}
		responseJSON, _ := json.Marshal(response)
		m.Respond(responseJSON)
	}
}

type ExerciseLaunchRequest struct {
	UserID     string `json:"user_id"`
	ExerciseID string `json:"exercise_id"`
}

type ExerciseLaunchResponse struct {
	SessionID string           `json:"session_id"`
	Exercise  *models.Exercise `json:"exercise"`
}

func HandleExerciseLaunch(es *services.ExerciseService) nats.MsgHandler {
	return func(m *nats.Msg) {
		var req ExerciseLaunchRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			m.Respond([]byte(`{"error": "Invalid request"}`))
			return
		}

		exercise, sessionID, err := es.LaunchExercise(req.UserID, req.ExerciseID)
		if err != nil {
			m.Respond([]byte(`{"error": "Failed to launch exercise"}`))
			return
		}

		response := ExerciseLaunchResponse{
			SessionID: sessionID,
			Exercise:  exercise,
		}
		responseJSON, _ := json.Marshal(response)
		m.Respond(responseJSON)
	}
}

type PoolExercisesRequest struct {
	PoolID string `json:"pool_id"`
}

type PoolExercisesResponse struct {
	Exercises []*models.Exercise `json:"exercises"`
}

func HandlePoolExercises(es *services.ExerciseService) nats.MsgHandler {
	return func(m *nats.Msg) {
		var req PoolExercisesRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			m.Respond([]byte(`{"error": "Invalid request"}`))
			return
		}

		exercises, err := es.GetPoolExercises(req.PoolID)
		if err != nil {
			m.Respond([]byte(`{"error": "Failed to get pool exercises"}`))
			return
		}

		response := PoolExercisesResponse{Exercises: exercises}
		responseJSON, _ := json.Marshal(response)
		m.Respond(responseJSON)
	}
}

type ExerciseGetRequest struct {
	ExerciseID string `json:"exercise_id"`
}

type ExerciseGetResponse struct {
	Exercise *models.Exercise `json:"exercise"`
}

func HandleExerciseGet(es *services.ExerciseService) nats.MsgHandler {
	return func(m *nats.Msg) {
		var req ExerciseGetRequest
		if err := json.Unmarshal(m.Data, &req); err != nil {
			m.Respond([]byte(`{"error": "Invalid request"}`))
			return
		}

		exercise, err := es.GetExercise(req.ExerciseID)
		if err != nil {
			m.Respond([]byte(`{"error": "Exercise not found"}`))
			return
		}

		response := ExerciseGetResponse{Exercise: exercise}
		responseJSON, _ := json.Marshal(response)
		m.Respond(responseJSON)
	}
}
