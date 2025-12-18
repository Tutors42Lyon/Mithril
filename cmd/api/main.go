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
	poolsRepo := repository.NewPoolRepository(db)
	exerciseRepo := repository.NewExerciseRepository(db)

	nc, err := nats.Connect(env.NatsUrl)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer nc.Close()

	messaging.LoadWorker(nc, userRepo, poolsRepo, exerciseRepo)
	r := gin.Default()

	authHandler := handlers.NewAuthHandler(nc, env)
	userHandler := handlers.NewUserHandler(nc)
	poolHandler := handlers.NewPoolHandler(nc)
	exerciceHandler := handlers.NewExerciseHandler(nc)

	r.GET("/auth/login/init", authHandler.LoginInit)

	// ex: /auth/poll?session_id=xxxxx...
	r.GET("/auth/poll", authHandler.PollLogin)
	r.GET("/callback", authHandler.CallBack)

	//ex: /users/role/pnaessen  body :  "role": "admin"
	r.PATCH("/users/role/:username", userHandler.UpdateRole)
	//ex: /users/info/pnaessen || /users/info/cassie
	r.GET("/users/info/:username", userHandler.GetUserInfo)

	r.GET("/exercises/pools", poolHandler.GetPoolsInfo)
	r.PUT("/pool/add", poolHandler.AddPool)
	//get all exercises from a pool
	r.GET("pool/:poolname", exerciceHandler.GetExercisesFromPool)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("cannot run the serv %v", err)
	}
}
