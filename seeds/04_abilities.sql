INSERT INTO abilities (code, name, description, icon, damage, heal, mp_cost) VALUES
-- Империя
('sith_warrior_basic_attack', 'Нападение', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/1562632956_1963234752.png', 15, 0, 10),
('sith_inquisitor_basic_attack', 'Удар мечем', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/1562632956_1963234752.png', 15, 0, 10),
('imperial_agent_basic_attack', 'Снайперский выстрел', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/3665256898_900988075.png', 15, 0, 10),
('bounty_hunter_basic_attack', 'Серия выстрелов', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/rapidshots.png', 15, 0, 10),
-- Республика
('jedi_knight_basic_attack', 'Удар', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/2837517449_3771294176.png', 15, 0, 10),
('jedi_consular_basic_attack', 'Удар мечем', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/1720182283_1756577701.png', 15, 0, 10),
('smuggler_basic_attack', 'Серия выстрелов', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/2174459481_1534558053.png', 15, 0, 10),
('trooper_basic_attack', 'Бластерный выстрел', 'Наносит %d% урона оружием выбранной цели.', 'images/abilities/hammershot.png', 15, 0, 10)
ON CONFLICT (code) DO NOTHING;
