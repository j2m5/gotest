BEGIN;

CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    planet_id INT NOT NULL REFERENCES planets(id),
    code VARCHAR(64) NOT NULL CHECK (code ~ '^[a-z][a-z0-9_]*$'),
    name VARCHAR(100) NOT NULL CHECK (char_length(btrim(name)) > 0),
    description TEXT NOT NULL DEFAULT '',
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    recommended_level SMALLINT NOT NULL REFERENCES levels(level),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    UNIQUE (planet_id,code),
    UNIQUE (planet_id,position)
);
CREATE TRIGGER locations_updated_at BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();

-- Upgrades may contain characters but no planet seed yet. Bootstrap only their
-- required capital content; fresh databases receive the catalog through seeds.
INSERT INTO planets(code,name,name_en,core_faction_id,controlling_faction_id)
SELECT CASE f.code WHEN 'galactic_republic' THEN 'coruscant' ELSE 'dromund_kaas' END,
       CASE f.code WHEN 'galactic_republic' THEN 'Корускант' ELSE 'Дромунд-Каас' END,
       CASE f.code WHEN 'galactic_republic' THEN 'Coruscant' ELSE 'Dromund Kaas' END,f.id,f.id
FROM factions f WHERE f.code IN ('galactic_republic','sith_empire')
ON CONFLICT (code) DO NOTHING;

INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,1,1 FROM (VALUES
 ('coruscant','lower_residential_levels','Нижние жилые уровни','Жилые переходы и тесные переулки под сияющими башнями столицы. Здесь повседневная жизнь мегаполиса соседствует с опасностями нижних уровней.'),
 ('dromund_kaas','kaas_city_outskirts','Окраины Каас-Сити','Дороги у границы имперской столицы уходят под густые кроны джунглей. За освещёнными постами начинаются заросли и влажные ущелья.')
) AS v(planet_code,code,name,description) JOIN planets p ON p.code=v.planet_code
WHERE EXISTS(SELECT 1 FROM levels WHERE level=1)
ON CONFLICT (planet_id,code) DO NOTHING;

ALTER TABLE characters ADD COLUMN location_id INT REFERENCES locations(id);
UPDATE characters c SET location_id=(
    SELECT l.id FROM factions f JOIN planets p ON p.code=CASE f.code
        WHEN 'galactic_republic' THEN 'coruscant' WHEN 'sith_empire' THEN 'dromund_kaas' END
    JOIN locations l ON l.planet_id=p.id WHERE f.id=c.faction_id AND p.core_faction_id=f.id
    ORDER BY l.recommended_level,l.position,l.id LIMIT 1
);
-- Fail atomically if existing custom factions have no configured starting world.
ALTER TABLE characters ALTER COLUMN location_id SET NOT NULL;
CREATE INDEX characters_location_idx ON characters(location_id);

CREATE FUNCTION set_character_start_location() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.location_id IS NULL THEN
        SELECT l.id INTO NEW.location_id FROM factions f JOIN planets p ON p.code=CASE f.code
            WHEN 'galactic_republic' THEN 'coruscant' WHEN 'sith_empire' THEN 'dromund_kaas' END
        JOIN locations l ON l.planet_id=p.id WHERE f.id=NEW.faction_id AND p.core_faction_id=f.id
        ORDER BY l.recommended_level,l.position,l.id LIMIT 1;
        IF NEW.location_id IS NULL THEN
            RAISE EXCEPTION 'character starting location is missing; seed capital locations'
                USING ERRCODE='23503', CONSTRAINT='characters_start_location_required';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER characters_start_location BEFORE INSERT ON characters
    FOR EACH ROW EXECUTE FUNCTION set_character_start_location();

COMMIT;
