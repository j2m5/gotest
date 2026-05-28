CREATE TABLE factions (
      id SERIAL PRIMARY KEY,
      name VARCHAR(32) NOT NULL,
      alias VARCHAR(32) NOT NULL,
      description TEXT,
      icon VARCHAR(255),
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);