package database

import (
	"fmt"
	"log"

	"github.com/Tutors42Lyon/Mithril/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(env *config.Env) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		env.DBHost,
		env.DBUser,
		env.DBPassword,
		env.DBName,
		env.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}
