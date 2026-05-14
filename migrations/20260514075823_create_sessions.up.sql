CREATE TABLE sessions (
                          id SERIAL PRIMARY KEY,

                          user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

                          token TEXT NOT NULL UNIQUE,

                          created_at TIMESTAMP NOT NULL DEFAULT NOW(),

                          updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);