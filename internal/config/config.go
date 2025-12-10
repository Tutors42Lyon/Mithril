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
}

func LoadEnv() (*Env, error) {

	_ = godotenv.Load()

	env := &Env{
		ClientID:     os.Getenv("API_42_UID"),
		ClientSecret: os.Getenv("API_42_SEC"),
		RedirectURL:  os.Getenv("CALL_BACK"),
		NatsUrl:      os.ExpandEnv("NATS_URL"),
	}

	if env.ClientID == "" || env.ClientSecret == "" || env.RedirectURL == "" {
		return nil, fmt.Errorf("missing required environment variable(s): API_42_UID, API_42_SEC, CALL_BACK")
	}
	return env, nil
}
