CREATE TABLE classes (
     id SERIAL PRIMARY KEY,
     faction_id INT NOT NULL REFERENCES factions(id) ON DELETE CASCADE,
     name VARCHAR(32) NOT NULL,
     alias VARCHAR(32),
     description TEXT,
     icon VARCHAR(255),
     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);