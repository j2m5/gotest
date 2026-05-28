CREATE TABLE class_abilities (
     class_id INT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
     ability_id INT NOT NULL REFERENCES abilities(id) ON DELETE CASCADE,
     PRIMARY KEY (class_id, ability_id)
);