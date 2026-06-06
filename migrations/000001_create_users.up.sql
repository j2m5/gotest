CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    role_id SMALLINT NOT NULL DEFAULT 1,
    character_id BIGINT,
    name VARCHAR(32),
    email VARCHAR(32) NOT NULL UNIQUE,
    login VARCHAR(32) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email_verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);