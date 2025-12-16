package messaging

import (
	"encoding/json"
	"log"

	"github.com/Tutors42Lyon/Mithril/internal/models"
	repository "github.com/Tutors42Lyon/Mithril/internal/repositories"

	"github.com/nats-io/nats.go"
)

func respondError(m *nats.Msg, code string, message string, httpStatus int) {
	resp := struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		HTTPStatus int    `json:"http_status,omitempty"`
	}{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshal error response: %v", err)
		return
	}
	if err := m.Respond(respBytes); err != nil {
		log.Printf("Error responding to NATS message: %v", err)
	}
}

func LoadWorker(nc *nats.Conn, userRepo *repository.UserRepository, poolRepo *repository.PoolRepository) {

	_, err := nc.Subscribe("user.login", HandleUserLogin(userRepo))

	if err != nil {
		log.Fatalf("Error Subscribe to NATS: %v", err)
	}

	_, err = nc.Subscribe("user.update_role", HandleUserUpdateRole(userRepo))

	if err != nil {
		log.Fatalf("Error Subscribe to NATS: %v", err)
	}

	_, err = nc.Subscribe("user.get_info", HandleUserInfo(userRepo))
	if err != nil {
		log.Fatalf("Error Subscribe to NATS: %v", err)
	}

	_, err = nc.Subscribe("pools.get_info", HandlePoolsInfo(poolRepo))
	if err != nil {
		log.Fatalf("Error Subscribe to NATS: %v", err)
	}
}

func HandleUserLogin(userRepo *repository.UserRepository) nats.MsgHandler {
	return func(m *nats.Msg) {
		var user models.UserMessage

		if err := json.Unmarshal(m.Data, &user); err != nil {
			log.Printf("Error Subscribe Unmarshal %v", err)
			respondError(m, "bad_request", "invalid payload", 400)
			return
		}

		if err := userRepo.CreateUser(&user); err != nil {
			log.Printf("Error saving to DB: %v", err)
			respondError(m, "db_error", "unable to save user", 500)
			return
		}

		resp := models.UserMessage{
			Username: "Test HandleUserLogin resp",
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshal response: %v", err)
			respondError(m, "internal_error", "marshal failed", 500)
			return
		}

		if err := m.Respond(respBytes); err != nil {
			log.Printf("Error responding to NATS message: %v", err)
			return
		}
	}
}

func HandleUserUpdateRole(userRepo *repository.UserRepository) nats.MsgHandler {

	return func(m *nats.Msg) {
		var req struct {
			Username string `json:"username"`
			Role     string `json:"role"`
		}

		if err := json.Unmarshal(m.Data, &req); err != nil {
			log.Printf("Error Subscribe Unmarshal %v", err)
			respondError(m, "bad_request", "invalid payload", 400)
			return
		}

		if req.Role != "admin" && req.Role != "student" && req.Role != "instructor" {
			log.Printf("Error: Invalid role provided")
			respondError(m, "invalid_role", "role must be admin|student|instructor", 400)
			return
		}

		if err := userRepo.UpdateUserRoleByUsername(req.Username, req.Role); err != nil {
			log.Printf("Error update role")
			respondError(m, "db_error", "unable to update role", 500)
			return
		}

		resp := models.UserMessage{
			Username: req.Username,
			Role:     req.Role,
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshal response: %v", err)
			respondError(m, "internal_error", "marshal failed", 500)
			return
		}

		if err := m.Respond(respBytes); err != nil {
			log.Printf("Error responding to NATS message: %v", err)
			return
		}
	}
}

func HandleUserInfo(userRepo *repository.UserRepository) nats.MsgHandler {

	return func(m *nats.Msg) {
		var req struct {
			Username string `json:"username"`
		}

		if err := json.Unmarshal(m.Data, &req); err != nil {
			log.Printf("Error Subscribe Unmarshal %v", err)
			respondError(m, "bad_request", "invalid payload", 400)
			return
		}

		if req.Username == "" {
			log.Printf("Error: empty username in request")
			respondError(m, "bad_request", "username required", 400)
			return
		}

		userInfo, err := userRepo.GetByUsername(req.Username)
		if err != nil {
			log.Printf("Error fetching user info: %v", err)
			respondError(m, "db_error", "user not found", 404)
			return
		}

		respBytes, err := json.Marshal(userInfo)
		if err != nil {
			log.Printf("Error marshal response: %v", err)
			respondError(m, "internal_error", "marshal failed", 500)
			return
		}

		if err := m.Respond(respBytes); err != nil {
			log.Printf("Error responding to NATS message: %v", err)
			return
		}

	}
}

func HandlePoolsInfo(poolRepo *repository.PoolRepository) nats.MsgHandler {

	return func(m *nats.Msg) {
		pools, err := poolRepo.GetAll()
		if err != nil {
			m.Respond([]byte(`{"error":"db error"}`))
			return
		}

		data, _ := json.Marshal(pools)
		m.Respond(data)
	}
}