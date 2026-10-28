CREATE TABLE chat_messages (
 id BIGSERIAL PRIMARY KEY,
 user_id INT NOT NULL REFERENCES users(id),
 character_id INT NOT NULL REFERENCES characters(id),
 author_name VARCHAR(32) NOT NULL,
 author_role VARCHAR(32) NOT NULL CHECK(author_role IN ('user','moderator','administrator')),
 body TEXT NOT NULL CHECK(char_length(btrim(body)) BETWEEN 1 AND 1000),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 deleted_at TIMESTAMPTZ,
 deleted_by_user_id INT REFERENCES users(id)
);
CREATE INDEX chat_messages_user_created_idx ON chat_messages(user_id,created_at DESC);
ALTER TABLE moderation_actions ADD CONSTRAINT moderation_actions_message_id_fkey FOREIGN KEY(message_id) REFERENCES chat_messages(id);
