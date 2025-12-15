---
layout: default
title: Project Structure
parent: Development
nav_order: 3
---

# Project Structure

Overview of the Mithril project structure.

## Directory Layout

## Backend Structure

```
mithril_content/
└── pools/                       # Root container for all content
    └── rust-beginner/           # [POOL FOLDER] - The collection name (slug)
        ├── pool.yaml            # [POOL CONFIG] - Meta info for the whole pool
        ├── 01-forging-start/    # [EXERCISE FOLDER] - Specific exercise
        │   ├── exercise.yaml    # [EXERCISE CONFIG] - Meta info for this task
        │   ├── README.md        # The instructions for the user (optional but rec.)
        │   ├── src/             # The starter code for the user
        │   │   └── main.rs
        │   └── tests/           # [TEST FOLDER] - Validation logic
        │       └── test_case.rs
        └── 02-mining-loops/
            ├── exercise.yaml
            ├── src/
            └── tests/

```

## Frontend Structure

## Documentation


```
mithril-backend/
│
├── cmd/                          # Service entry points
│   ├── api/
│   │   └── main.go              # API Gateway startup
│   ├── grading/
│   │   └── main.go              # Grading Service startup
│   └── migrate/
│       └── main.go              # Database migrations
│
├── internal/                     # Private application code
│   ├── services/                # Business logic
│   │   ├── auth.go              # Authentication (42 OAuth, sessions)
│   │   ├── exercise.go          # Exercise retrieval & caching
│   │   ├── submission.go        # Submission handling
│   │   └── pool.go              # Exercise pool management
│   │
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth.go              # Login, logout, callback endpoints
│   │   ├── exercises.go         # Exercise list, detail endpoints
│   │   ├── submissions.go       # Submit, check status endpoints
│   │   ├── pools.go             # Pool listing endpoints
│   │   └── middleware.go        # Auth, logging, errors
│   │
│   ├── models/                  # Data structures (Go structs)
│   │   ├── user.go              # User struct
│   │   ├── exercise.go          # Exercise, ExerciseContent structs
│   │   ├── pool.go              # ExercisePool struct
│   │   ├── attempt.go           # ExerciseAttempt struct
│   │   ├── session.go           # Session struct
│   │   └── grading.go           # GradingJob, GradingResult structs
│   │
│   ├── database/                # PostgreSQL interactions
│   │   ├── db.go                # Connection setup & pooling
│   │   ├── queries.go           # Query builders (optional)
│   │   ├── connection.go        # Connection management
│   │   └── migrations/          # SQL migration files
│   │       ├── 001_init.sql     # Create tables
│   │       ├── 002_indexes.sql  # Add indexes
│   │       └── 003_seed.sql     # Test data
│   │
│   ├── cache/                   # Redis interactions
│   │   ├── redis.go             # Redis client setup
│   │   ├── session.go           # Session caching
│   │   ├── exercise.go          # Exercise caching
│   │   └── keys.go              # Redis key constants
│   │
│   ├── messaging/               # NATS message queue
│   │   ├── nats.go              # NATS client setup
│   │   ├── publisher.go         # Publish functions
│   │   └── subscriber.go        # Subscribe & consume functions
│   │
│   ├── grading/                 # Grading logic
│   │   ├── executor.go          # Execute user code
│   │   ├── tester.go            # Run test cases
│   │   ├── compiler.go          # Compile C code
│   │   └── errors.go            # Error handling
│   │
│   ├── yaml/                    # YAML exercise loading
│   │   ├── loader.go            # Load YAML files
│   │   ├── parser.go            # Parse YAML structure
│   │   └── validator.go         # Validate YAML format
│   │
│   ├── auth42/                  # 42 Intra OAuth
│   │   ├── client.go            # 42 API HTTP client
│   │   ├── oauth.go             # OAuth flow
│   │   └── user.go              # Fetch user from 42
│   │
│   └── config/                  # Configuration
│       └── config.go            # Load env & config files
│
├── pkg/                         # Public packages (reusable)
│   └── (empty for MVP)
│
├── test/                        # Integration tests
│   ├── integration_test.go
│   ├── fixtures/
│   │   └── test_data.json
│   └── README.md
│
├── scripts/                     # Utility scripts
│   ├── setup.sh                 # Project setup
│   ├── seed-db.sh              # Seed test data
│   └── test.sh                 # Run tests
│
├── config/                      # Configuration files
│   ├── .env.example            # Env var template
│   └── config.yaml             # (Optional)
│
├── docker/                      # Separate Dockerfiles for each service
│   ├── api.Dockerfile          # API Gateway image
│   ├── grading.Dockerfile      # Grading Service image
│   └── migrations.Dockerfile   # Migrations tool image
│
├── exercises/                   # Exercise YAML files
│   ├── pool_1_c_basics/
│   │   ├── hello_world.yaml
│   │   ├── arrays.yaml
│   │   └── loops.yaml
│   ├── pool_2_algorithms/
│   │   ├── bubble_sort.yaml
│   │   ├── binary_search.yaml
│   │   └── quicksort.yaml
│   └── pool_3_git/
│       ├── basic_commands.yaml
│       └── branching.yaml
│
```

