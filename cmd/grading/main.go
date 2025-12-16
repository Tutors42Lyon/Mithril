package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

const natsURL = "nats://nats:4222"
const requestTimeout = 5 * time.Second

type GradingService struct {
	NC *nats.Conn
}

func main() {

	log.Println("Grading Service started")

	nc, err := nats.Connect(natsURL,
		nats.Name("Grading-Service-Worker"),
		nats.ErrorHandler(func(conn *nats.Conn, sub *nats.Subscription, err error) {
			log.Printf("Async error NATS : %v", err)
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			log.Fatalf("NATS Connection closed. End of program.")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			log.Printf("Disconneted from NATS : %v. Try to reconnect...", err)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			log.Printf("Reconnected to NATS. New server : %s", conn.ConnectedUrl())
		}),
	)
	if err != nil {
		log.Fatalf("Failed to connect to NATS : %v", err)
	}
	defer nc.Close()

	service := &GradingService{NC: nc}

	_, err = service.NC.Subscribe("grading.*.*.submit", service.handleSubmission)
	if err != nil {
		log.Fatalf("Failed to subscribe NATS : %v", err)
	}

	log.Println("Grading Service is running and listening...")

	waitForTermination()
}

func (s *GradingService) handleSubmission(m *nats.Msg) {

	subject := strings.Split(m.Subject, ".")
	if len(subject) < 4 {
		log.Printf("Message ingored : subject misformated (%s)", m.Subject)
		return
	}

	clientId := subject[1]
	exerciseId := subject[2]

	log.Printf("New submission : Client %s, Exercise %s", clientId, exerciseId)

	go s.processGrading(m.Data, exerciseId, clientId)
}

func (s *GradingService) processGrading(data []byte, exerciseId string, clientId string) {

	workerSubject := "worker." + exerciseId + ".grade"
	resp, err := s.NC.Request(workerSubject, data, requestTimeout)

	if err != nil {
		errorMessage := ""
		if err != nats.ErrTimeout {
			errorMessage = "Worker timedout"
			log.Printf("Timeout Request for %s (Client %s) : %v", exerciseId, clientId, err)
		} else if err == nats.ErrNoResponders {
			errorMessage = "No Worker available"
			log.Printf("No Worker for %s (Client %s) : %v", exerciseId, clientId, err)
		} else {
			errorMessage = "NATS communication error"
			log.Printf("Unexpected NATS error for %s (Client %s) : %v", exerciseId, clientId, err)
		}

		errorResultSubject := "grading." + clientId + "." + exerciseId + ".error"
		s.NC.Publish(errorResultSubject, []byte(errorMessage))
		return
	}

	resultSubject := "grading." + clientId + "." + exerciseId + ".result"
	s.NC.Publish(resultSubject, resp.Data)
	log.Printf("Results published for Client %s, Exercise %s", clientId, exerciseId)

	// add status update
}

func waitForTermination() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
