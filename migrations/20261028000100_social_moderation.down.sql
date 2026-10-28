DROP TABLE moderation_actions;
DROP TABLE user_bans;
ALTER TABLE users DROP CONSTRAINT users_role_id_fkey;
DROP TABLE account_roles;
