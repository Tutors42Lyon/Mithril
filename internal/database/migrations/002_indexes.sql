CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

CREATE INDEX idx_pools_category ON exercise_pools(category);
CREATE INDEX idx_pools_is_published ON exercise_pools(is_published);

CREATE INDEX idx_exercises_pool_id ON exercises(pool_id);
CREATE INDEX idx_exercises_order ON exercises(pool_id, order_index);
