package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	NatsUrl      string
	DBHost       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBPort       string
	JwtSecret    string
}

func LoadEnv() (*Env, error) {

	_ = godotenv.Load()

	env := &Env{
		ClientID:     os.Getenv("API_42_UID"),
		ClientSecret: os.Getenv("API_42_SEC"),
		RedirectURL:  os.Getenv("CALL_BACK"),
		NatsUrl:      os.ExpandEnv("NATS_URL"),
		DBHost:       os.Getenv("DB_HOST"),
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		DBPort:       os.Getenv("DB_PORT"),
		JwtSecret:    os.Getenv("JWT_SECRET"),
	}

	if env.ClientID == "" || env.ClientSecret == "" || env.RedirectURL == "" || env.NatsUrl == "" {
		return nil, fmt.Errorf("missing required environment variable(s): API_42_UID, API_42_SEC, CALL_BACK")
	}

	if env.DBHost == "" || env.DBUser == "" || env.DBPassword == "" || env.DBName == "" || env.DBPort == "" {
		return nil, fmt.Errorf("missing required database environment variable(s): check DB_HOST, DB_USER, etc.")
	}
	return env, nil
}
