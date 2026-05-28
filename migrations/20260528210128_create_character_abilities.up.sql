CREATE TABLE character_abilities (
     character_id INT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
     ability_id INT NOT NULL REFERENCES abilities(id) ON DELETE CASCADE,
     PRIMARY KEY (character_id, ability_id)
);