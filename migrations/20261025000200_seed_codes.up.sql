BEGIN;

LOCK TABLE factions, classes, abilities IN ACCESS EXCLUSIVE MODE;
ALTER TABLE factions ADD COLUMN code VARCHAR(64);
ALTER TABLE classes ADD COLUMN code VARCHAR(64);
ALTER TABLE abilities ADD COLUMN code VARCHAR(64);

-- Adopt existing reference data without replacing IDs or touching game balance.
UPDATE factions SET code = CASE name
    WHEN 'Sith Empire' THEN 'sith_empire'
    WHEN 'Galactic Republic' THEN 'galactic_republic'
    ELSE 'legacy_faction_' || id::text
END;

UPDATE classes c SET code = seed.code
FROM (VALUES
    ('sith_warrior','Sith Warrior','sith_empire'),
    ('sith_inquisitor','Sith Inquisitor','sith_empire'),
    ('imperial_agent','Imperial Agent','sith_empire'),
    ('bounty_hunter','Bounty Hunter','sith_empire'),
    ('jedi_knight','Jedi Knight','galactic_republic'),
    ('jedi_consular','Jedi Consular','galactic_republic'),
    ('smuggler','Smuggler','galactic_republic'),
    ('trooper','Trooper','galactic_republic')
) AS seed(code,name,faction_code), factions f
WHERE c.name=seed.name AND c.faction_id=f.id AND f.code=seed.faction_code;
UPDATE classes SET code='legacy_class_' || id::text WHERE code IS NULL;

-- Display names repeat across classes; the original icon disambiguates them.
UPDATE abilities a SET code = seed.code
FROM (VALUES
    ('sith_warrior_basic_attack','Нападение','images/abilities/1562632956_1963234752.png'),
    ('sith_inquisitor_basic_attack','Удар мечем','images/abilities/1562632956_1963234752.png'),
    ('imperial_agent_basic_attack','Снайперский выстрел','images/abilities/3665256898_900988075.png'),
    ('bounty_hunter_basic_attack','Серия выстрелов','images/abilities/rapidshots.png'),
    ('jedi_knight_basic_attack','Удар','images/abilities/2837517449_3771294176.png'),
    ('jedi_consular_basic_attack','Удар мечем','images/abilities/1720182283_1756577701.png'),
    ('smuggler_basic_attack','Серия выстрелов','images/abilities/2174459481_1534558053.png'),
    ('trooper_basic_attack','Бластерный выстрел','images/abilities/hammershot.png')
) AS seed(code,name,icon)
WHERE a.name=seed.name AND a.icon=seed.icon;
UPDATE abilities SET code='legacy_ability_' || id::text WHERE code IS NULL;

-- Do not guess which duplicate owns existing character/class references.
DO $$
BEGIN
    IF EXISTS (SELECT code FROM factions GROUP BY code HAVING count(*)>1) THEN
        RAISE EXCEPTION 'seed codes: duplicate legacy factions; resolve duplicates before migrating';
    END IF;
    IF EXISTS (SELECT code FROM classes GROUP BY code HAVING count(*)>1) THEN
        RAISE EXCEPTION 'seed codes: duplicate legacy classes; resolve duplicates before migrating';
    END IF;
    IF EXISTS (SELECT code FROM abilities GROUP BY code HAVING count(*)>1) THEN
        RAISE EXCEPTION 'seed codes: duplicate legacy abilities; resolve duplicates before migrating';
    END IF;
END $$;

ALTER TABLE factions
    ALTER COLUMN code SET NOT NULL,
    ADD CONSTRAINT factions_code_key UNIQUE(code),
    ADD CONSTRAINT factions_code_format CHECK(code ~ '^[a-z][a-z0-9_]*$');
ALTER TABLE classes
    ALTER COLUMN code SET NOT NULL,
    ADD CONSTRAINT classes_code_key UNIQUE(code),
    ADD CONSTRAINT classes_code_format CHECK(code ~ '^[a-z][a-z0-9_]*$');
ALTER TABLE abilities
    ALTER COLUMN code SET NOT NULL,
    ADD CONSTRAINT abilities_code_key UNIQUE(code),
    ADD CONSTRAINT abilities_code_format CHECK(code ~ '^[a-z][a-z0-9_]*$');

COMMIT;
