CREATE TABLE class_roles (
     class_id INT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
     role VARCHAR(16) NOT NULL,
     PRIMARY KEY (class_id, role)
);