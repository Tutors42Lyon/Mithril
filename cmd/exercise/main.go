package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Tutors42Lyon/Mithril/internal/cache"
	"github.com/Tutors42Lyon/Mithril/internal/messaging"
	"github.com/Tutors42Lyon/Mithril/internal/services"
	"github.com/Tutors42Lyon/Mithril/internal/yaml"
	"github.com/nats-io/nats.go"
)

func main() {
	// 1. Load configuration (exercise service only needs Redis and NATS)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379" // Default
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if redisDBStr := os.Getenv("REDIS_DB"); redisDBStr != "" {
		if parsedDB, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = parsedDB
		}
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222" // Default
	}

	log.Printf("Starting exercise service with Redis: %s, NATS: %s", redisURL, natsURL)

	// 2. Connect to Redis
	redisClient, err := cache.NewRedisClient(redisURL, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	exerciseCache := cache.NewExerciseCache(redisClient)

	// 3. Initialize YAML loader
	loader := yaml.NewLoader("/app/exercises")

	// 4. Initial load: scan and cache all exercises
	log.Println("Loading exercises from YAML...")
	pools, exercises, err := loader.LoadAll()
	if err != nil {
		log.Printf("Warning: Some exercises failed to load: %v", err)
	}
	log.Printf("Loaded %d pools and %d exercises", len(pools), len(exercises))

	// 5. Populate Redis cache
	poolIDs := []string{}
	for poolID, pool := range pools {
		if err := exerciseCache.SetPool(pool, 1*time.Hour); err != nil {
			log.Printf("Error caching pool %s: %v", poolID, err)
		}
		poolIDs = append(poolIDs, poolID)
	}
	exerciseCache.SetPoolsList(poolIDs, 30*time.Minute)

	for exerciseID, exercise := range exercises {
		if err := exerciseCache.SetExercise(exercise, 1*time.Hour); err != nil {
			log.Printf("Error caching exercise %s: %v", exerciseID, err)
		}
	}
	log.Println("Cache populated")

	// 6. Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS")

	// 7. Setup NATS subscribers
	exerciseService := services.NewExerciseService(exerciseCache, loader, nc)
	messaging.LoadExerciseSubscribers(nc, exerciseService)
	log.Println("NATS subscribers loaded")

	log.Println("Exercise service is running...")
	select {} // Block forever
}
