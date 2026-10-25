WITH seed(class_code, role) AS (
VALUES
    ('sith_warrior', 'tank'),
    ('sith_warrior', 'dd'),
    ('sith_inquisitor', 'tank'),
    ('sith_inquisitor', 'dd'),
    ('sith_inquisitor', 'healer'),
    ('imperial_agent', 'dd'),
    ('imperial_agent', 'healer'),
    ('bounty_hunter', 'tank'),
    ('bounty_hunter', 'dd'),
    ('bounty_hunter', 'healer'),
    ('jedi_knight', 'tank'),
    ('jedi_knight', 'dd'),
    ('jedi_consular', 'tank'),
    ('jedi_consular', 'dd'),
    ('jedi_consular', 'healer'),
    ('smuggler', 'dd'),
    ('smuggler', 'healer'),
    ('trooper', 'tank'),
    ('trooper', 'dd'),
    ('trooper', 'healer')
)
INSERT INTO class_roles (class_id, role)
SELECT (SELECT id FROM classes WHERE code=seed.class_code), seed.role
FROM seed
ON CONFLICT (class_id, role) DO NOTHING;
