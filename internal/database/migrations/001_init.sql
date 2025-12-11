DROP DATABASE postgres;
ALTER DATABASE template0 IS_TEMPLATE FALSE;
DROP DATABASE template0;
ALTER DATABASE template1 IS_TEMPLATE FALSE;
DROP DATABASE template1;

\c mithril;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL,
    intra_id INT UNIQUE,
    school_year VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

CREATE TABLE exercise_pools (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL, 
    description TEXT,
    is_published BOOLEAN DEFAULT false,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    pool_id INT REFERENCES exercise_pools(id) ON DELETE CASCADE,
    yaml_file_path VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    points INT DEFAULT 0,
    order_index INT,
    chapter TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);