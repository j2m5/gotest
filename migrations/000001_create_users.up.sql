CREATE TABLE users (
                       id SERIAL PRIMARY KEY,

                       username VARCHAR(32) NOT NULL UNIQUE,
                       password TEXT NOT NULL,

                       role VARCHAR(16) NOT NULL DEFAULT 'user',

                       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);