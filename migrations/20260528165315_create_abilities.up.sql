CREATE TABLE abilities (
       id SERIAL PRIMARY KEY,
       name VARCHAR(64) NOT NULL,
       description TEXT,
       icon VARCHAR(255),
       damage INT NOT NULL DEFAULT 0,
       heal INT NOT NULL DEFAULT 0,
       mp_cost INT NOT NULL DEFAULT 0,
       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);