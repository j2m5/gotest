INSERT INTO factions(id,name,alias) VALUES
    (41,'Sith Empire','Empire'), (42,'Galactic Republic','Republic'), (99,'Custom faction','Custom');
INSERT INTO classes(id,faction_id,name) VALUES
    (101,41,'Sith Inquisitor'), (102,42,'Jedi Consular'), (199,99,'Custom class');
-- Identical display names represent different abilities.
INSERT INTO abilities(id,name,icon,damage) VALUES
    (201,'Удар мечем','images/abilities/1562632956_1963234752.png',77),
    (202,'Удар мечем','images/abilities/1720182283_1756577701.png',88),
    (299,'Custom ability',NULL,99);
INSERT INTO class_roles(class_id,role) VALUES (101,'tank');
INSERT INTO class_abilities(class_id,ability_id) VALUES (101,201),(102,202);
INSERT INTO users(id,email,login,password) VALUES (1,'fixture@example.test','fixture','unused');
INSERT INTO characters(id,user_id,faction_id,class_id,name) VALUES (1,1,41,101,'Existing character');
INSERT INTO character_abilities(character_id,ability_id) VALUES (1,201);
SELECT setval('factions_id_seq',99);
SELECT setval('classes_id_seq',199);
SELECT setval('abilities_id_seq',299);
