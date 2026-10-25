BEGIN;

-- Prevent concurrent writes from invalidating the diagnostics before constraints
-- are installed. Existing rows are validated, never repaired or deleted here.
LOCK TABLE characters, class_roles, levels IN ACCESS EXCLUSIVE MODE;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM levels WHERE level <= 0 OR experience_required < 0) THEN
        RAISE EXCEPTION 'character stats: levels catalog requires positive levels and nonnegative experience; repair existing rows before migrating';
    END IF;
    IF EXISTS (
        SELECT 1 FROM characters c WHERE NOT EXISTS (
            SELECT 1 FROM class_roles r WHERE r.class_id=c.class_id AND r.role=c.role
        )
    ) THEN
        RAISE EXCEPTION 'character stats: class roles must include each existing character role; repair the catalog or characters before migrating';
    END IF;
    IF EXISTS (
        SELECT 1 FROM characters c WHERE NOT EXISTS (SELECT 1 FROM levels l WHERE l.level=c.level)
    ) THEN
        RAISE EXCEPTION 'character stats: character levels must exist in the levels catalog; repair existing references before migrating';
    END IF;
    IF EXISTS (
        SELECT 1 FROM characters WHERE experience < 0
            OR hp <= 0 OR hp_current < 0 OR hp_current > hp
            OR mp < 0 OR mp_current < 0 OR mp_current > mp
            OR mastery < 0 OR endurance < 0 OR power < 0 OR accuracy < 0
            OR alacrity < 0 OR critical < 0 OR defense < 0 OR shield < 0 OR absorption < 0
    ) THEN
        RAISE EXCEPTION 'character stats: invalid numeric values exist; repair experience, resource bounds and nonnegative attributes before migrating';
    END IF;
END;
$$;

ALTER TABLE levels
    ADD CONSTRAINT levels_level_positive_check CHECK (level > 0),
    ADD CONSTRAINT levels_experience_nonnegative_check CHECK (experience_required >= 0);

ALTER TABLE characters
    ADD CONSTRAINT characters_class_role_fkey FOREIGN KEY (class_id, role) REFERENCES class_roles(class_id, role),
    ADD CONSTRAINT characters_level_fkey FOREIGN KEY (level) REFERENCES levels(level),
    ADD CONSTRAINT characters_experience_nonnegative_check CHECK (experience >= 0),
    ADD CONSTRAINT characters_hp_bounds_check CHECK (hp > 0 AND hp_current >= 0 AND hp_current <= hp),
    ADD CONSTRAINT characters_mp_bounds_check CHECK (mp >= 0 AND mp_current >= 0 AND mp_current <= mp),
    ADD CONSTRAINT characters_attributes_nonnegative_check CHECK (
        mastery >= 0 AND endurance >= 0 AND power >= 0 AND accuracy >= 0
        AND alacrity >= 0 AND critical >= 0 AND defense >= 0 AND shield >= 0 AND absorption >= 0
    );

-- PostgreSQL does not automatically index the referencing side of foreign keys.
CREATE INDEX characters_class_role_idx ON characters(class_id, role);
CREATE INDEX characters_level_idx ON characters(level);

COMMIT;
