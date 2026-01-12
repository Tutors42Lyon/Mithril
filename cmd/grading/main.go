package main

import (
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

var nc *nats.Conn

func main() {

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("grading.*.*.submit", handleSubmission)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Grading Service is running and listening...")

	select {}
}

func handleSubmission(m *nats.Msg) {

	subject := strings.Split(m.Subject, ".")
	clientId := subject[1]
	exerciseId := subject[2]

	go processGrading(m.Data, exerciseId, clientId)
}

func processGrading(data []byte, exerciseId string, clientId string) {

	resp, err := nc.Request("worker."+exerciseId+".grade", data, 5*time.Second)

	if err != nil {
		log.Fatal(err)
	}

	//edit resp.data

	nc.Publish("grading."+clientId+"."+exerciseId+".result", resp.Data)
}
