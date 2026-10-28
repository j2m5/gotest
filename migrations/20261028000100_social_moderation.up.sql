CREATE TABLE account_roles (id SMALLINT PRIMARY KEY, name VARCHAR(32) NOT NULL UNIQUE);
INSERT INTO account_roles VALUES (1,'user'),(2,'moderator'),(3,'administrator');
ALTER TABLE users ADD CONSTRAINT users_role_id_fkey FOREIGN KEY(role_id) REFERENCES account_roles(id);
CREATE TABLE user_bans (
 id BIGSERIAL PRIMARY KEY,
 user_id INT NOT NULL REFERENCES users(id),
 scope VARCHAR(16) NOT NULL CHECK(scope IN ('account','chat','forum')),
 reason TEXT NOT NULL CHECK(char_length(btrim(reason)) BETWEEN 3 AND 500),
 created_by_user_id INT NOT NULL REFERENCES users(id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 expires_at TIMESTAMPTZ,
 revoked_at TIMESTAMPTZ,
 revoked_by_user_id INT REFERENCES users(id),
 CHECK ((expires_at IS NULL AND scope='account') OR (expires_at IS NOT NULL AND expires_at-created_at BETWEEN interval '1 hour' AND interval '168 hours' AND mod(extract(epoch FROM expires_at-created_at),3600)=0)),
 CHECK ((revoked_at IS NULL)=(revoked_by_user_id IS NULL))
);
CREATE INDEX user_bans_active_idx ON user_bans(user_id,scope) WHERE revoked_at IS NULL;
CREATE TABLE moderation_actions (
 id BIGSERIAL PRIMARY KEY,
 actor_user_id INT REFERENCES users(id),
 user_id INT NOT NULL REFERENCES users(id),
 scope VARCHAR(16) NOT NULL CHECK(scope IN ('account','chat','forum')),
 action VARCHAR(32) NOT NULL CHECK(action IN ('ban','revoke','delete_message','role_change')),
 reason TEXT NOT NULL CHECK(char_length(btrim(reason)) BETWEEN 3 AND 500),
 ban_id BIGINT REFERENCES user_bans(id),
 message_id BIGINT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
