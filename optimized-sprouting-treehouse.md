# Exercise Content Handling System - Implementation Plan

## Overview

This plan implements an exercise content handling system with:
- **New Exercise Service** (microservice) - Parses YAML exercise files and manages Redis cache
- **Redis Caching Layer** - Fast exercise retrieval with fallback to YAML
- **WebSocket Gateway** - Real-time client communication bridged to NATS
- **NATS Integration** - Leverages existing message broker for inter-service communication

## User Requirements

Real-time delivery of:
- Exercise pools information
- Exercise details and metadata
- Progress & status updates during exercise work
- Grading results
- Exercise details when launching an exercise

---

## Architecture Decision

**Pattern**: NATS + WebSocket Gateway (chosen by user)
- Clients connect via WebSocket to API service
- WebSocket handler bridges messages to/from NATS topics
- Exercise service subscribes to NATS topics and responds
- Scales horizontally with multiple service replicas

---

## Implementation Phases

### Phase 1: Redis Infrastructure & Data Models

#### 1.1 Enable Redis in Docker

**File**: `/home/cassie/Documents/Mithril/docker-compose.yml`
- Uncomment Redis service (lines 22-33)
- Uncomment Redis dependency in API service (lines 83-84)

#### 1.2 Add Redis Client Library

**File**: `/home/cassie/Documents/Mithril/go.mod`
```bash
go get github.com/redis/go-redis/v9
go get github.com/gorilla/websocket
```

#### 1.3 Update Config Structure

**File**: `/home/cassie/Documents/Mithril/internal/config/config.go`

Add Redis fields to `Env` struct:
```go
type Env struct {
    // ... existing fields
    RedisURL      string
    RedisPassword string
    RedisDB       int
}
```

Update `LoadEnv()` to read:
- `REDIS_URL` (already in .env)
- `REDIS_PASSWORD` (already in .env)
- `REDIS_DB` (already in .env)

#### 1.4 Define Data Models

**File**: `/home/cassie/Documents/Mithril/internal/models/pool.go` (currently 1 line)

```go
type Pool struct {
    ID          string    `json:"id" yaml:"slug"`
    Name        string    `json:"name" yaml:"name"`
    Theme       string    `json:"theme" yaml:"theme"`
    Description string    `json:"description" yaml:"description"`
    Difficulty  string    `json:"difficulty" yaml:"difficulty"`
    Tags        []string  `json:"tags" yaml:"tags"`
    Maintainer  string    `json:"maintainer" yaml:"maintainer"`
    Exercises   []string  `json:"exercises"`
}
```

**File**: `/home/cassie/Documents/Mithril/internal/models/exercise.go` (currently 1 line)

```go
type Exercise struct {
    ID          string            `json:"id" yaml:"id"`
    PoolID      string            `json:"pool_id"`
    Title       string            `json:"title" yaml:"title"`
    Type        string            `json:"type" yaml:"type"` // code, input, qcm, text
    Language    string            `json:"language" yaml:"language"`
    Build       *BuildConfig      `json:"build,omitempty" yaml:"build"`
    Tests       []TestCase        `json:"tests" yaml:"tests"`
    Validation  *ValidationRules  `json:"validation,omitempty" yaml:"validation"`
    Scoring     *ScoringConfig    `json:"scoring,omitempty" yaml:"scoring"`
}

type BuildConfig struct {
    Command string `json:"command" yaml:"command"`
    Timeout int    `json:"timeout" yaml:"timeout"`
}

type TestCase struct {
    Name           string `json:"name" yaml:"name"`
    Run            string `json:"run" yaml:"run"`
    Input          string `json:"input" yaml:"input"`
    ExpectedOutput string `json:"expected_output" yaml:"expected_output"`
    Timeout        int    `json:"timeout" yaml:"timeout"`
}

type ValidationRules struct {
    CheckValgrind      bool     `json:"check_valgrind" yaml:"check_valgrind"`
    AllowedFunctions   []string `json:"allowed_functions" yaml:"allowed_functions"`
    ForbiddenFunctions []string `json:"forbidden_functions" yaml:"forbidden_functions"`
}

type ScoringConfig struct {
    Compilation   int `json:"compilation" yaml:"compilation"`
    PerTest       int `json:"per_test" yaml:"per_test"`
    ValgrindClean int `json:"valgrind_clean" yaml:"valgrind_clean"`
}
```

#### 1.5 Implement Redis Cache Layer

**File**: `/home/cassie/Documents/Mithril/internal/cache/redis.go` (currently 1 line)

```go
package cache

import (
    "context"
    "github.com/redis/go-redis/v9"
    "time"
)

type RedisClient struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisClient(url string, password string, db int) (*RedisClient, error) {
    opts, err := redis.ParseURL(url)
    if err != nil {
        return nil, err
    }

    opts.Password = password
    opts.DB = db
    opts.PoolSize = 10

    client := redis.NewClient(opts)
    ctx := context.Background()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    return &RedisClient{client: client, ctx: ctx}, nil
}

func (rc *RedisClient) Get(key string) (string, error)
func (rc *RedisClient) Set(key string, value string, ttl time.Duration) error
func (rc *RedisClient) Delete(key string) error
func (rc *RedisClient) Exists(key string) (bool, error)
func (rc *RedisClient) GetList(key string) ([]string, error)
func (rc *RedisClient) AddToList(key string, values ...string) error
```

**File**: `/home/cassie/Documents/Mithril/internal/cache/keys.go` (currently 1 line)

```go
package cache

const (
    KeyPrefixPool     = "mithril:pool:"
    KeyPrefixExercise = "mithril:exercise:"
    KeyPoolsList      = "mithril:pools:list"
    KeyPoolExercises  = "mithril:pool:%s:exercises" // Format with pool_id
)

func PoolKey(poolID string) string {
    return KeyPrefixPool + poolID
}

func ExerciseKey(exerciseID string) string {
    return KeyPrefixExercise + exerciseID
}

func PoolExercisesKey(poolID string) string {
    return fmt.Sprintf(KeyPoolExercises, poolID)
}
```

**File**: `/home/cassie/Documents/Mithril/internal/cache/exercise.go` (currently 1 line)

```go
package cache

import (
    "encoding/json"
    "time"
    "github.com/Tutors42Lyon/Mithril/internal/models"
)

type ExerciseCache struct {
    redis *RedisClient
}

func NewExerciseCache(redis *RedisClient) *ExerciseCache {
    return &ExerciseCache{redis: redis}
}

// Pool operations
func (ec *ExerciseCache) SetPool(pool *models.Pool, ttl time.Duration) error
func (ec *ExerciseCache) GetPool(poolID string) (*models.Pool, error)
func (ec *ExerciseCache) ListPools() ([]string, error)
func (ec *ExerciseCache) SetPoolsList(poolIDs []string, ttl time.Duration) error

// Exercise operations
func (ec *ExerciseCache) SetExercise(exercise *models.Exercise, ttl time.Duration) error
func (ec *ExerciseCache) GetExercise(exerciseID string) (*models.Exercise, error)
func (ec *ExerciseCache) GetPoolExercises(poolID string) ([]string, error)
func (ec *ExerciseCache) SetPoolExercises(poolID string, exerciseIDs []string, ttl time.Duration) error

// Invalidation
func (ec *ExerciseCache) InvalidatePool(poolID string) error
func (ec *ExerciseCache) InvalidateExercise(exerciseID string) error
```

**Cache Strategy**:
- TTL: 1 hour (from .env CACHE_TTL=3600)
- Cache-first with YAML fallback on miss
- Manual invalidation on file changes

---

### Phase 2: YAML Parsing System

#### 2.1 Implement YAML Parser

**File**: `/home/cassie/Documents/Mithril/internal/yaml/parser.go` (currently 1 line)

Pattern: Improve upon existing worker service approach (`cmd/worker/main.go:94-132`)

```go
package yaml

import (
    "os"
    "gopkg.in/yaml.v2"
    "github.com/Tutors42Lyon/Mithril/internal/models"
)

type Parser struct {
    exercisesDir string
}

func NewParser(exercisesDir string) *Parser {
    return &Parser{exercisesDir: exercisesDir}
}

func (p *Parser) ParsePool(poolPath string) (*models.Pool, error) {
    data, err := os.ReadFile(poolPath)
    if err != nil {
        return nil, err
    }

    var pool models.Pool
    if err := yaml.Unmarshal(data, &pool); err != nil {
        return nil, fmt.Errorf("failed to parse pool %s: %w", poolPath, err)
    }

    return &pool, nil
}

func (p *Parser) ParseExercise(exercisePath string, poolID string) (*models.Exercise, error) {
    data, err := os.ReadFile(exercisePath)
    if err != nil {
        return nil, err
    }

    var exercise models.Exercise
    if err := yaml.Unmarshal(data, &exercise); err != nil {
        return nil, fmt.Errorf("failed to parse exercise %s: %w", exercisePath, err)
    }

    exercise.PoolID = poolID
    return &exercise, nil
}
```

#### 2.2 Implement YAML Loader

**File**: `/home/cassie/Documents/Mithril/internal/yaml/loader.go` (currently 1 line)

```go
package yaml

import (
    "os"
    "path/filepath"
    "strings"
    "github.com/Tutors42Lyon/Mithril/internal/models"
)

type Loader struct {
    parser       *Parser
    exercisesDir string
}

func NewLoader(exercisesDir string) *Loader {
    return &Loader{
        parser:       NewParser(exercisesDir),
        exercisesDir: exercisesDir,
    }
}

// LoadAll scans exercises directory and loads all pools and exercises
func (l *Loader) LoadAll() (map[string]*models.Pool, map[string]*models.Exercise, error) {
    pools := make(map[string]*models.Pool)
    exercises := make(map[string]*models.Exercise)
    errors := []error{}

    // Walk exercises directory
    err := filepath.Walk(l.exercisesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Find pool.yaml files
        if !info.IsDir() && info.Name() == "pool.yaml" {
            pool, err := l.parser.ParsePool(path)
            if err != nil {
                log.Printf("Error parsing pool %s: %v", path, err)
                errors = append(errors, err)
                return nil // Continue walking
            }
            pools[pool.ID] = pool

            // Load exercises in this pool
            poolDir := filepath.Dir(path)
            l.loadPoolExercises(poolDir, pool.ID, exercises, &errors)
        }

        return nil
    })

    if err != nil {
        return nil, nil, err
    }

    if len(errors) > 0 {
        return pools, exercises, fmt.Errorf("loaded with %d errors", len(errors))
    }

    return pools, exercises, nil
}

func (l *Loader) loadPoolExercises(poolDir string, poolID string, exercises map[string]*models.Exercise, errors *[]error) {
    // Scan for exercise.yaml files in subdirectories
    filepath.Walk(poolDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && info.Name() == "exercise.yaml" {
            exercise, err := l.parser.ParseExercise(path, poolID)
            if err != nil {
                log.Printf("Error parsing exercise %s: %v", path, err)
                *errors = append(*errors, err)
                return nil
            }
            exercises[exercise.ID] = exercise
        }
        return nil
    })
}

// LoadPool loads a specific pool and its exercises
func (l *Loader) LoadPool(poolID string) (*models.Pool, []*models.Exercise, error)

// LoadExercise loads a specific exercise
func (l *Loader) LoadExercise(exerciseID string) (*models.Exercise, error)
```

**Error Handling**: Log malformed YAML files but continue loading others. Return aggregated errors.

---

### Phase 3: Exercise Microservice

#### 3.1 Create Exercise Service Main

**File**: `/home/cassie/Documents/Mithril/cmd/exercise/main.go` (new file)

Pattern: Follow existing service structure from `cmd/api/main.go` and `cmd/worker/main.go`

```go
package main

import (
    "log"
    "time"
    "github.com/Tutors42Lyon/Mithril/internal/config"
    "github.com/Tutors42Lyon/Mithril/internal/cache"
    "github.com/Tutors42Lyon/Mithril/internal/yaml"
    "github.com/Tutors42Lyon/Mithril/internal/services"
    "github.com/Tutors42Lyon/Mithril/internal/messaging"
    "github.com/nats-io/nats.go"
)

func main() {
    // 1. Load configuration
    env, err := config.LoadEnv()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 2. Connect to Redis
    redisClient, err := cache.NewRedisClient(env.RedisURL, env.RedisPassword, env.RedisDB)
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
    nc, err := nats.Connect(env.NatsUrl)
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
```

#### 3.2 Create Service Layer

**File**: `/home/cassie/Documents/Mithril/internal/services/exercise.go` (currently 1 line)

```go
package services

import (
    "github.com/Tutors42Lyon/Mithril/internal/cache"
    "github.com/Tutors42Lyon/Mithril/internal/yaml"
    "github.com/Tutors42Lyon/Mithril/internal/models"
    "github.com/nats-io/nats.go"
)

type ExerciseService struct {
    cache    *cache.ExerciseCache
    loader   *yaml.Loader
    natsConn *nats.Conn
}

func NewExerciseService(cache *cache.ExerciseCache, loader *yaml.Loader, nc *nats.Conn) *ExerciseService {
    return &ExerciseService{
        cache:    cache,
        loader:   loader,
        natsConn: nc,
    }
}

// GetPoolList retrieves all pools (cache-first)
func (es *ExerciseService) GetPoolList() ([]*models.Pool, error) {
    poolIDs, err := es.cache.ListPools()
    if err != nil || len(poolIDs) == 0 {
        // Cache miss - reload from YAML
        pools, _, err := es.loader.LoadAll()
        return poolsMapToSlice(pools), err
    }

    // Get each pool from cache
    pools := []*models.Pool{}
    for _, poolID := range poolIDs {
        pool, err := es.cache.GetPool(poolID)
        if err != nil {
            continue
        }
        pools = append(pools, pool)
    }
    return pools, nil
}

// GetPool retrieves a specific pool
func (es *ExerciseService) GetPool(poolID string) (*models.Pool, error) {
    pool, err := es.cache.GetPool(poolID)
    if err != nil {
        // Cache miss - load from YAML
        pool, _, err = es.loader.LoadPool(poolID)
        if err != nil {
            return nil, err
        }
        // Cache for future
        es.cache.SetPool(pool, 1*time.Hour)
    }
    return pool, nil
}

// GetExercise retrieves a specific exercise
func (es *ExerciseService) GetExercise(exerciseID string) (*models.Exercise, error) {
    exercise, err := es.cache.GetExercise(exerciseID)
    if err != nil {
        // Cache miss - load from YAML
        exercise, err = es.loader.LoadExercise(exerciseID)
        if err != nil {
            return nil, err
        }
        es.cache.SetExercise(exercise, 1*time.Hour)
    }
    return exercise, nil
}

// GetPoolExercises retrieves all exercises in a pool
func (es *ExerciseService) GetPoolExercises(poolID string) ([]*models.Exercise, error)

// LaunchExercise creates a session for user to start an exercise
func (es *ExerciseService) LaunchExercise(userID string, exerciseID string) (*models.Exercise, string, error) {
    exercise, err := es.GetExercise(exerciseID)
    if err != nil {
        return nil, "", err
    }

    sessionID := generateSessionID()

    // Publish exercise.{userID}.status with "started" status
    es.publishStatus(userID, exerciseID, "started")

    return exercise, sessionID, nil
}

// ReloadExercise reloads exercise from YAML and updates cache
func (es *ExerciseService) ReloadExercise(poolID, exerciseID string) error

func (es *ExerciseService) publishStatus(userID, exerciseID, status string) {
    // Publish to NATS topic for WebSocket forwarding
}
```

#### 3.3 Create NATS Message Handlers

**File**: `/home/cassie/Documents/Mithril/internal/messaging/exercise.go` (new file)

Pattern: Follow existing pattern from `internal/messaging/subscriber.go`

```go
package messaging

import (
    "encoding/json"
    "github.com/nats-io/nats.go"
    "github.com/Tutors42Lyon/Mithril/internal/services"
)

func LoadExerciseSubscribers(nc *nats.Conn, exerciseService *services.ExerciseService) {
    // Use queue groups for load balancing (like worker service)
    nc.QueueSubscribe("exercise.pool.list", "exercise.service", HandlePoolList(exerciseService))
    nc.QueueSubscribe("exercise.pool.get", "exercise.service", HandlePoolGet(exerciseService))
    nc.QueueSubscribe("exercise.pool.exercises", "exercise.service", HandlePoolExercises(exerciseService))
    nc.QueueSubscribe("exercise.get", "exercise.service", HandleExerciseGet(exerciseService))
    nc.QueueSubscribe("exercise.launch", "exercise.service", HandleExerciseLaunch(exerciseService))
}

type PoolListResponse struct {
    Pools []*models.Pool `json:"pools"`
}

func HandlePoolList(es *services.ExerciseService) nats.MsgHandler {
    return func(m *nats.Msg) {
        pools, err := es.GetPoolList()
        if err != nil {
            m.Respond([]byte(`{"error": "Failed to get pool list"}`))
            return
        }

        response := PoolListResponse{Pools: pools}
        responseJSON, _ := json.Marshal(response)
        m.Respond(responseJSON)
    }
}

type PoolGetRequest struct {
    PoolID string `json:"pool_id"`
}

type PoolGetResponse struct {
    Pool *models.Pool `json:"pool"`
}

func HandlePoolGet(es *services.ExerciseService) nats.MsgHandler {
    return func(m *nats.Msg) {
        var req PoolGetRequest
        if err := json.Unmarshal(m.Data, &req); err != nil {
            m.Respond([]byte(`{"error": "Invalid request"}`))
            return
        }

        pool, err := es.GetPool(req.PoolID)
        if err != nil {
            m.Respond([]byte(`{"error": "Pool not found"}`))
            return
        }

        response := PoolGetResponse{Pool: pool}
        responseJSON, _ := json.Marshal(response)
        m.Respond(responseJSON)
    }
}

type ExerciseLaunchRequest struct {
    UserID     string `json:"user_id"`
    ExerciseID string `json:"exercise_id"`
}

type ExerciseLaunchResponse struct {
    SessionID string           `json:"session_id"`
    Exercise  *models.Exercise `json:"exercise"`
}

func HandleExerciseLaunch(es *services.ExerciseService) nats.MsgHandler {
    return func(m *nats.Msg) {
        var req ExerciseLaunchRequest
        if err := json.Unmarshal(m.Data, &req); err != nil {
            m.Respond([]byte(`{"error": "Invalid request"}`))
            return
        }

        exercise, sessionID, err := es.LaunchExercise(req.UserID, req.ExerciseID)
        if err != nil {
            m.Respond([]byte(`{"error": "Failed to launch exercise"}`))
            return
        }

        response := ExerciseLaunchResponse{
            SessionID: sessionID,
            Exercise:  exercise,
        }
        responseJSON, _ := json.Marshal(response)
        m.Respond(responseJSON)
    }
}

// Similar handlers for HandlePoolExercises and HandleExerciseGet
```

#### 3.4 Create Dockerfile

**File**: `/home/cassie/Documents/Mithril/docker/exercise.Dockerfile` (new file)

Pattern: Follow existing Dockerfiles (api.Dockerfile, worker.Dockerfile)

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o exercise ./cmd/exercise

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/exercise .
COPY --from=builder /app/exercises ./exercises

CMD ["./exercise"]
```

#### 3.5 Add to Docker Compose

**File**: `/home/cassie/Documents/Mithril/docker-compose.yml`

Add after grading service (around line 105):

```yaml
  exercise:
    build:
      context: .
      dockerfile: docker/exercise.Dockerfile
    container_name: mithril-exercise
    environment:
      REDIS_URL: "redis://redis:6379"
      NATS_URL: "nats://nats:4222"
      LOG_LEVEL: "debug"
    depends_on:
      redis:
        condition: service_healthy
      nats:
        condition: service_started
    volumes:
      - ./exercises:/app/exercises:ro  # Read-only mount for exercise files
    networks:
      - mithril-network
    restart: unless-stopped
```

---

### Phase 4: WebSocket Gateway

#### 4.1 Create WebSocket Handler

**File**: `/home/cassie/Documents/Mithril/internal/handlers/websocket.go` (new file)

```go
package handlers

import (
    "encoding/json"
    "log"
    "sync"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/nats-io/nats.go"
    "github.com/Tutors42Lyon/Mithril/internal/utils"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // TODO: Add proper origin checking
    },
}

type WebSocketHandler struct {
    natsConn *nats.Conn
    clients  map[string]*WebSocketClient
    mu       sync.RWMutex
}

type WebSocketClient struct {
    conn          *websocket.Conn
    userID        string
    subscriptions map[string]*nats.Subscription
    send          chan []byte
    handler       *WebSocketHandler
}

func NewWebSocketHandler(nc *nats.Conn) *WebSocketHandler {
    return &WebSocketHandler{
        natsConn: nc,
        clients:  make(map[string]*WebSocketClient),
    }
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (wsh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    // 1. Extract and validate JWT token
    token := c.Query("token")
    if token == "" {
        token = c.GetHeader("Authorization")
    }

    claims, err := utils.ValidateJWT(token)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid token"})
        return
    }

    userID := claims["user_id"].(string)

    // 2. Upgrade to WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }

    // 3. Create client
    client := &WebSocketClient{
        conn:          conn,
        userID:        userID,
        subscriptions: make(map[string]*nats.Subscription),
        send:          make(chan []byte, 256),
        handler:       wsh,
    }

    wsh.mu.Lock()
    wsh.clients[userID] = client
    wsh.mu.Unlock()

    log.Printf("WebSocket connected: user=%s", userID)

    // 4. Start read/write pumps
    go client.writePump()
    go client.readPump()
}

// WebSocket message types
type WSMessage struct {
    Type      string          `json:"type"` // "subscribe", "unsubscribe", "request"
    Subject   string          `json:"subject,omitempty"`
    Topics    []string        `json:"topics,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
    RequestID string          `json:"request_id,omitempty"`
}

type WSResponse struct {
    Type      string          `json:"type"` // "response", "event", "error"
    Subject   string          `json:"subject,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
    RequestID string          `json:"request_id,omitempty"`
    Error     string          `json:"error,omitempty"`
}

func (client *WebSocketClient) readPump() {
    defer func() {
        client.Close()
    }()

    for {
        _, message, err := client.conn.ReadMessage()
        if err != nil {
            break
        }

        var wsMsg WSMessage
        if err := json.Unmarshal(message, &wsMsg); err != nil {
            client.SendError("Invalid message format")
            continue
        }

        switch wsMsg.Type {
        case "subscribe":
            client.Subscribe(wsMsg.Topics)
        case "unsubscribe":
            client.Unsubscribe(wsMsg.Topics)
        case "request":
            client.ForwardRequest(wsMsg)
        }
    }
}

func (client *WebSocketClient) writePump() {
    for message := range client.send {
        if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
            return
        }
    }
}

// Subscribe client to NATS topics and forward messages to WebSocket
func (client *WebSocketClient) Subscribe(topics []string) {
    for _, topic := range topics {
        // Replace {user_id} placeholder with actual userID
        topic = strings.ReplaceAll(topic, "{user_id}", client.userID)

        sub, err := client.handler.natsConn.Subscribe(topic, func(m *nats.Msg) {
            // Forward NATS message to WebSocket
            response := WSResponse{
                Type:    "event",
                Subject: m.Subject,
                Data:    json.RawMessage(m.Data),
            }
            responseJSON, _ := json.Marshal(response)
            client.send <- responseJSON
        })

        if err != nil {
            log.Printf("Subscribe error for %s: %v", topic, err)
            continue
        }

        client.subscriptions[topic] = sub
        log.Printf("User %s subscribed to %s", client.userID, topic)
    }
}

func (client *WebSocketClient) Unsubscribe(topics []string) {
    for _, topic := range topics {
        topic = strings.ReplaceAll(topic, "{user_id}", client.userID)

        if sub, exists := client.subscriptions[topic]; exists {
            sub.Unsubscribe()
            delete(client.subscriptions, topic)
            log.Printf("User %s unsubscribed from %s", client.userID, topic)
        }
    }
}

// ForwardRequest forwards WebSocket request to NATS and returns response
func (client *WebSocketClient) ForwardRequest(wsMsg WSMessage) {
    // Make NATS request
    resp, err := client.handler.natsConn.Request(wsMsg.Subject, wsMsg.Data, 5*time.Second)

    response := WSResponse{
        Type:      "response",
        RequestID: wsMsg.RequestID,
    }

    if err != nil {
        response.Error = "Request timeout or error"
    } else {
        response.Data = json.RawMessage(resp.Data)
    }

    responseJSON, _ := json.Marshal(response)
    client.send <- responseJSON
}

func (client *WebSocketClient) SendError(message string) {
    response := WSResponse{
        Type:  "error",
        Error: message,
    }
    responseJSON, _ := json.Marshal(response)
    client.send <- responseJSON
}

func (client *WebSocketClient) Close() {
    // Unsubscribe from all NATS topics
    for _, sub := range client.subscriptions {
        sub.Unsubscribe()
    }

    // Remove from clients map
    client.handler.mu.Lock()
    delete(client.handler.clients, client.userID)
    client.handler.mu.Unlock()

    close(client.send)
    client.conn.Close()

    log.Printf("WebSocket disconnected: user=%s", client.userID)
}
```

#### 4.2 Integrate WebSocket into API Service

**File**: `/home/cassie/Documents/Mithril/cmd/api/main.go`

Add after line 31 (after NATS connection):

```go
// Initialize WebSocket handler
wsHandler := handlers.NewWebSocketHandler(nc)
```

Add before line 50 (before r.Run):

```go
// WebSocket endpoint
r.GET("/ws", wsHandler.HandleWebSocket)
```

---

### Phase 5: Integration & Real-time Updates

#### 5.1 Integrate Grading Results with WebSocket

**File**: `/home/cassie/Documents/Mithril/cmd/grading/main.go`

Modify grading result publishing to include status updates:

```go
// After receiving grading result from worker
resp, err := nc.Request("worker."+exerciseId+".grade", data, 5*time.Second)

// Publish to existing grading result topic
nc.Publish("grading."+clientId+"."+exerciseId+".result", resp.Data)

// Also publish status update for WebSocket
statusMsg := map[string]interface{}{
    "user_id":     clientId,
    "exercise_id": exerciseId,
    "status":      "completed",
    "timestamp":   time.Now(),
}
statusJSON, _ := json.Marshal(statusMsg)
nc.Publish("exercise."+clientId+".status", statusJSON)
```

---

## NATS Topics Structure

### Request-Reply Topics (Synchronous)
- `exercise.pool.list` - Get all pools
- `exercise.pool.get` - Get specific pool by ID
- `exercise.pool.exercises` - Get exercises in a pool
- `exercise.get` - Get specific exercise by ID
- `exercise.launch` - Launch exercise for user

### Publish-Subscribe Topics (Real-time Updates)
- `exercise.{user_id}.progress` - User progress updates
- `exercise.{user_id}.status` - Exercise status changes
- `grading.{user_id}.{exercise_id}.result` - Grading results (existing)

### Integration with Existing Topics
- `worker.{exercise_id}.grade` - Existing grading queue (unchanged)
- `grading.{client_id}.{exercise_id}.submit` - Existing submission topic (unchanged)

---

## Redis Cache Key Design

```
mithril:pools:list                     → LIST of pool IDs
mithril:pool:{pool_id}                 → JSON of pool metadata
mithril:pool:{pool_id}:exercises       → LIST of exercise IDs
mithril:exercise:{exercise_id}         → JSON of exercise data
```

**Example**:
```
mithril:pools:list → ["regex-beginner", "rust-basics"]
mithril:pool:regex-beginner → {"id":"regex-beginner","name":"Regex Beginner",...}
mithril:pool:regex-beginner:exercises → ["ex01_match", "ex02_groups"]
mithril:exercise:ex01_match → {"id":"ex01_match","title":"Basic Matching",...}
```

---

## WebSocket Protocol

### Client → Server Messages

**Subscribe to topics**:
```json
{
  "type": "subscribe",
  "topics": ["exercise.{user_id}.progress", "exercise.{user_id}.status"]
}
```

**Make request**:
```json
{
  "type": "request",
  "subject": "exercise.pool.list",
  "data": {},
  "request_id": "uuid"
}
```

**Launch exercise**:
```json
{
  "type": "request",
  "subject": "exercise.launch",
  "data": {"user_id": "12345", "exercise_id": "ex01_hello_world"},
  "request_id": "uuid"
}
```

### Server → Client Messages

**Response to request**:
```json
{
  "type": "response",
  "request_id": "uuid",
  "data": {"pools": [...]}
}
```

**Real-time event**:
```json
{
  "type": "event",
  "subject": "exercise.12345.progress",
  "data": {"exercise_id": "ex01", "progress": 60, "tests_passed": 3, "tests_total": 5}
}
```

**Error**:
```json
{
  "type": "error",
  "error": "Authentication failed"
}
```

---

## Error Handling & Fallback Strategy

### Redis Connection Failures
- **Strategy**: Fallback to direct YAML loading
- **Implementation**: All cache.Get* methods check Redis error and call loader on failure

### YAML Parsing Errors
- **Strategy**: Skip malformed files, continue loading others
- **Implementation**: Log errors but don't fail entire load operation

### WebSocket Disconnections
- **Strategy**: Auto-reconnect with exponential backoff (client-side)
- **Implementation**: Clean up subscriptions on server disconnect

### NATS Message Timeouts
- **Strategy**: 5-second timeout with error response
- **Implementation**: Use `nc.Request(subject, data, 5*time.Second)`

---

## Implementation Order

1. **Phase 1**: Redis + Data Models (1-2 days)
   - Enable Redis, add dependencies
   - Implement cache layer
   - Define data models

2. **Phase 2**: YAML Parsing (1 day)
   - Implement parser and loader
   - Test with existing exercise files

3. **Phase 3**: Exercise Service (2 days)
   - Create microservice
   - Implement service layer
   - Add NATS subscribers
   - Docker integration

4. **Phase 4**: WebSocket Gateway (2 days)
   - Implement WebSocket handler
   - Add to API service
   - Test bidirectional communication

5. **Phase 5**: Integration (1 day)
   - Integrate grading with new topics
   - End-to-end testing
   - Documentation

---

## Testing Checklist

- [ ] Redis connection and cache operations
- [ ] YAML parsing for all exercise types
- [ ] Exercise service NATS request-reply
- [ ] WebSocket authentication
- [ ] WebSocket ↔ NATS message bridging
- [ ] Real-time progress updates
- [ ] Grading results forwarding
- [ ] Cache fallback when Redis is down
- [ ] Multiple concurrent WebSocket connections
- [ ] Exercise launch flow end-to-end

---

## Performance Targets

- Cache hit rate: >95% for exercise retrieval
- WebSocket latency: <50ms for message relay
- Exercise load time: <100ms (cached), <500ms (YAML)
- Support: 1000+ concurrent WebSocket connections
- NATS request-reply: <10ms avg latency

---

## Critical Files Summary

**New Files** (12 files):
1. `cmd/exercise/main.go` - Exercise service entry point
2. `internal/handlers/websocket.go` - WebSocket gateway
3. `internal/messaging/exercise.go` - NATS message handlers
4. `docker/exercise.Dockerfile` - Exercise service container

**Modified Files** (13 files):
5. `internal/models/pool.go` - Pool data structure
6. `internal/models/exercise.go` - Exercise data structure
7. `internal/cache/redis.go` - Redis client
8. `internal/cache/keys.go` - Cache key helpers
9. `internal/cache/exercise.go` - Exercise cache operations
10. `internal/yaml/parser.go` - YAML parsing
11. `internal/yaml/loader.go` - YAML loading and scanning
12. `internal/services/exercise.go` - Business logic
13. `internal/config/config.go` - Add Redis config fields
14. `cmd/api/main.go` - Add WebSocket route
15. `cmd/grading/main.go` - Add status publishing
16. `docker-compose.yml` - Uncomment Redis, add exercise service
17. `go.mod` - Add redis and websocket dependencies

---

## Notes

- Follow existing code patterns from `cmd/api/main.go` and `cmd/worker/main.go`
- Use queue groups for NATS subscribers: `nc.QueueSubscribe(topic, "exercise.service", handler)`
- Redis configuration already exists in `.env` - no changes needed
- WebSocket authentication uses existing JWT system from `internal/utils/jwt.go`
- Exercise service can be horizontally scaled (multiple replicas like worker service)
