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
			Username:   "Test HandleUserLogin resp",
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error marshal response: %v", err)
		}
		m.Respond(respBytes)
    }
}
