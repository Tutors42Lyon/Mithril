---
layout: default
title: Database
parent: Architecture
nav_order: 5
---

# Database

## Users Table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- 'student', 'instructor', 'admin'
    school_year INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

```


```sql
CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    theme VARCHAR(100), -- 'regex', 'bitwise', 'git', etc.
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    problem_statement TEXT NOT NULL, -- the actual problem
    yaml_file_path VARCHAR(255) NOT NULL, -- path to YAML file
    points INT DEFAULT 10,
    time_limit_minutes INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exercises_pool_id ON exercises(pool_id);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty_level);
CREATE INDEX idx_exercises_theme ON exercises(theme);

```

```sql
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'weekly_challenge', 'tournament', 'sprint', 'team_competition'
    status VARCHAR(50) DEFAULT 'upcoming', -- 'upcoming', 'active', 'ended'
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    rules TEXT,
    max_participants INT,
    is_team_based BOOLEAN DEFAULT false,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_start_date ON events(start_date);
```

