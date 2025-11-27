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
    role VARCHAR(50) NOT NULL, -- 'student', 'instructor', 'admin'
    intra_id INT UNIQUE,
    school_year INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

```


```sql
CREATE TABLE exercise_pools (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,  -- 'c_basics', 'algorithms', 'regex', 'git'
    description TEXT,
    is_published BOOLEAN DEFAULT false,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pools_category ON exercise_pools(category);
CREATE INDEX idx_pools_is_published ON exercise_pools(is_published);
```

```sql
CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    pool_id INT REFERENCES exercise_pools(id) ON DELETE CASCADE,
    yaml_file_path VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    points INT DEFAULT 0,
    order_index INT,  -- Position within pool (1, 2, 3...)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exercises_pool_id ON exercises(pool_id);
CREATE INDEX idx_exercises_order ON exercises(pool_id, order_index);

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
    max_participants INT,
    is_team_based BOOLEAN DEFAULT false,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_start_date ON events(start_date);
CREATE INDEX idx_events_created_by ON events(created_by);

```

```sql
CREATE TABLE event_exercises (
    id SERIAL PRIMARY KEY,
    event_id INT REFERENCES events(id) ON DELETE CASCADE,
    exercise_id INT REFERENCES exercises(id) ON DELETE CASCADE,
    points INT NOT NULL,
    time_limit_minutes INT,
    UNIQUE(event_id, exercise_id)
);

CREATE INDEX idx_event_exercises_event_id ON event_exercises(event_id);
CREATE INDEX idx_event_exercises_exercise_id ON event_exercises(exercise_id);

```

```sql
CREATE TABLE event_participants (
    id SERIAL PRIMARY KEY,
    event_id INT REFERENCES events(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    team_id INT REFERENCES teams(id) ON DELETE SET NULL,
    score INT DEFAULT 0,
    rank INT,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_submission TIMESTAMP,
    UNIQUE(event_id, user_id)
);

CREATE INDEX idx_event_participants_event_id ON event_participants(event_id);
CREATE INDEX idx_event_participants_user_id ON event_participants(user_id);
CREATE INDEX idx_event_participants_score ON event_participants(score DESC);

```


```sql
```
