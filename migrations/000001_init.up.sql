CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE expressions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expression TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'processing', 'done', 'error')),
    result DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE operations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_type TEXT NOT NULL CHECK(operation_type IN ('+', '-', '*', '/')),
    duration_ms INTEGER NOT NULL
);

CREATE INDEX idx_expressions_user_id ON expressions(user_id);
CREATE INDEX idx_operations_user_id ON operations(user_id);