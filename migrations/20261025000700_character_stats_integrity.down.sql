BEGIN;

LOCK TABLE characters, class_roles, levels IN ACCESS EXCLUSIVE MODE;

ALTER TABLE characters
    DROP CONSTRAINT characters_class_role_fkey,
    DROP CONSTRAINT characters_level_fkey,
    DROP CONSTRAINT characters_experience_nonnegative_check,
    DROP CONSTRAINT characters_hp_bounds_check,
    DROP CONSTRAINT characters_mp_bounds_check,
    DROP CONSTRAINT characters_attributes_nonnegative_check;

ALTER TABLE levels
    DROP CONSTRAINT levels_level_positive_check,
    DROP CONSTRAINT levels_experience_nonnegative_check;

DROP INDEX characters_class_role_idx;
DROP INDEX characters_level_idx;

COMMIT;
