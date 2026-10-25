WITH seed(class_code, ability_code) AS (
VALUES
    ('sith_warrior', 'sith_warrior_basic_attack'),
    ('sith_inquisitor', 'sith_inquisitor_basic_attack'),
    ('imperial_agent', 'imperial_agent_basic_attack'),
    ('bounty_hunter', 'bounty_hunter_basic_attack'),
    ('jedi_knight', 'jedi_knight_basic_attack'),
    ('jedi_consular', 'jedi_consular_basic_attack'),
    ('smuggler', 'smuggler_basic_attack'),
    ('trooper', 'trooper_basic_attack')
)
INSERT INTO class_abilities (class_id, ability_id)
SELECT (SELECT id FROM classes WHERE code=seed.class_code),
       (SELECT id FROM abilities WHERE code=seed.ability_code)
FROM seed
ON CONFLICT (class_id, ability_id) DO NOTHING;
