package messaging

import (
	"encoding/json"
	"log"
	"github.com/Tutors42Lyon/Mithril/internal/models"
	repository "github.com/Tutors42Lyon/Mithril/internal/repositories"

	"github.com/nats-io/nats.go"
)

func LoadWorker(nc *nats.Conn, userRepo *repository.UserRepository) {

	var user models.UserMessage

	_, err := nc.Subscribe("user.login", func(m *nats.Msg) {
		if err := json.Unmarshal(m.Data, &user); err != nil {
			log.Printf("Error Subscrite Unmarshal %v", err)
			return
		}

		if err := userRepo.CreateUser(&user); err != nil {
			log.Printf("Error saving to DB: %v", err)
			return
		}

		resp := models.UserMessage{
			Username:   user.Username,
			Email:      user.Email,
			Role:       "admin",
			IntraID:    user.IntraID,
			SchoolYear: user.SchoolYear,
			IsActive:   false,
			Db_id:      1,
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error marshal response: %v", err)
		}
		m.Respond(respBytes)
	})

	if err != nil {
		log.Fatalf("Error Subscribe to NATS: %v", err)
	}
}
