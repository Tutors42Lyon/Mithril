package messaging

import (
	"encoding/json"
	"log"

	"github.com/Tutors42Lyon/Mithril/internal/models"
	repository "github.com/Tutors42Lyon/Mithril/internal/repositories"

	"github.com/nats-io/nats.go"
)

func LoadWorker(nc *nats.Conn, userRepo *repository.UserRepository) {

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
}

func HandleUserLogin(userRepo *repository.UserRepository) nats.MsgHandler {
	return func(m *nats.Msg) {
		var user models.UserMessage

		if err := json.Unmarshal(m.Data, &user); err != nil {
			log.Printf("Error Subscribe Unmarshal %v", err)
			return
		}

		if err := userRepo.CreateUser(&user); err != nil {
			log.Printf("Error saving to DB: %v", err)
			return
		}

		resp := models.UserMessage{
			Username: "Test HandleUserLogin resp",
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshal response: %v", err)
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
			return
		}

		if req.Role != "admin" && req.Role != "student" && req.Role != "instructor" {
			log.Printf("Error: Invalid role provided")
			return
		}

		if err := userRepo.UpdateUserRoleByUsername(req.Username, req.Role); err != nil {
			log.Printf("Error update role")
			return
		}

		resp := models.UserMessage{
			Username: req.Username,
			Role:     req.Role,
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshal response: %v", err)
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
			return
		}

		if req.Username == "" {
            log.Printf("Error: empty username in request")
            return
        }

        userInfo, err := userRepo.GetByUsername(req.Username) // <-- renamed call
        if err != nil {
            log.Printf("Error fetching user info: %v", err)
            return
        }

        respBytes, err := json.Marshal(userInfo)
        if err != nil {
            log.Printf("Error marshal response: %v", err)
            return
        }

		if err := m.Respond(respBytes); err != nil {
            log.Printf("Error responding to NATS message: %v", err)
            return
        }

	}
}