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

	nc, err := nats.Connect(env.NatsUrl)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer nc.Close()

	messaging.LoadWorker(nc, userRepo)
	r := gin.Default()

	authHandler := handlers.NewAuthHandler(nc, env)
	userHandler := handlers.NewUserHandler(nc)

	// routes publique
	r.GET("/auth/login/init", authHandler.LoginInit)
	// ex: /auth/poll?session_id=xxxxx...
	r.GET("/auth/poll", authHandler.PollLogin)
	r.GET("/callback", authHandler.CallBack)

	usersGroups := r.Group("/")
	usersGroups.Use(handlers.AuthMiddleware())
	//ex: /users/info/pnaessen || /users/info/cassie
	usersGroups.GET("/users/info/:username", userHandler.GetUserInfo)
	//ex: /users/role/pnaessen  body :  "role": "admin"
	usersGroups.PATCH("/users/role/:username", userHandler.UpdateRole)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("cannot run the serv %v", err)
	}
}
