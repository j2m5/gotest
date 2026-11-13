BEGIN;
LOCK TABLE battles IN ACCESS EXCLUSIVE MODE;
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM battles) THEN
        RAISE EXCEPTION 'cannot roll back nonempty battles';
    END IF;
END $$;
DROP TABLE battles;
DROP FUNCTION guard_battle_update();
ALTER TABLE npc_spawns DROP COLUMN respawn_at;
COMMIT;
