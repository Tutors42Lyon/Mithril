package main

import (
	"log"

	"github.com/Tutors42Lyon/Mithril/internal/config"
	"github.com/Tutors42Lyon/Mithril/internal/database"
	"github.com/Tutors42Lyon/Mithril/internal/handlers"
	"github.com/Tutors42Lyon/Mithril/internal/messaging"
	repository "github.com/Tutors42Lyon/Mithril/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

func main() {

	env, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("cannot run the serv %v", err)
	}

	db := database.InitDB(env)

	userRepo := repository.NewUserRepository(db)

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer nc.Close()

	messaging.LoadWorker(nc, userRepo)
	r := gin.Default()

	authHandler := handlers.NewAuthHandler(nc, env)
	r.GET("/login", authHandler.Login)
	r.GET("/callback", authHandler.CallBack)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("cannot run the serv %v", err)
	}
}
