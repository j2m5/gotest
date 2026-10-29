CREATE TABLE forum_categories (
 id SERIAL PRIMARY KEY,
 code VARCHAR(32) NOT NULL UNIQUE,
 title VARCHAR(120) NOT NULL CHECK(char_length(btrim(title)) BETWEEN 1 AND 120),
 description TEXT NOT NULL DEFAULT '',
 position INT NOT NULL DEFAULT 0,
 create_role_id SMALLINT NOT NULL REFERENCES account_roles(id),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
INSERT INTO forum_categories(code,title,description,position,create_role_id) VALUES
 ('news','Новости','Новости и объявления проекта',1,3),
 ('general','Общение','Общий раздел для игроков',2,1),
 ('help','Помощь','Вопросы и помощь игрокам',3,1),
 ('suggestions','Предложения','Идеи по развитию игры',4,1);
CREATE TABLE forum_topics (
 id BIGSERIAL PRIMARY KEY,
 category_id INT NOT NULL REFERENCES forum_categories(id),
 user_id INT NOT NULL REFERENCES users(id),
 character_id INT NOT NULL,
 author_name VARCHAR(32) NOT NULL,
 title TEXT NOT NULL CHECK(char_length(btrim(title)) BETWEEN 3 AND 120),
 body TEXT NOT NULL CHECK(char_length(btrim(body)) BETWEEN 1 AND 10000),
 version INT NOT NULL DEFAULT 1 CHECK(version>0),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 edited_at TIMESTAMPTZ,
 closed_at TIMESTAMPTZ,
 pinned_at TIMESTAMPTZ,
 hidden_at TIMESTAMPTZ,
 hidden_by_user_id INT REFERENCES users(id),
 FOREIGN KEY(character_id,user_id) REFERENCES characters(id,user_id),
 CHECK((hidden_at IS NULL)=(hidden_by_user_id IS NULL))
);
CREATE INDEX forum_topics_category_page_idx ON forum_topics(category_id,(pinned_at IS NOT NULL) DESC,id DESC);
CREATE INDEX forum_topics_user_created_idx ON forum_topics(user_id,created_at DESC);
CREATE TABLE forum_posts (
 id BIGSERIAL PRIMARY KEY,
 topic_id BIGINT NOT NULL REFERENCES forum_topics(id),
 user_id INT NOT NULL REFERENCES users(id),
 character_id INT NOT NULL,
 author_name VARCHAR(32) NOT NULL,
 body TEXT NOT NULL CHECK(char_length(btrim(body)) BETWEEN 1 AND 10000),
 version INT NOT NULL DEFAULT 1 CHECK(version>0),
 created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
 edited_at TIMESTAMPTZ,
 hidden_at TIMESTAMPTZ,
 hidden_by_user_id INT REFERENCES users(id),
 FOREIGN KEY(character_id,user_id) REFERENCES characters(id,user_id),
 CHECK((hidden_at IS NULL)=(hidden_by_user_id IS NULL))
);
CREATE INDEX forum_posts_topic_page_idx ON forum_posts(topic_id,id);
CREATE INDEX forum_posts_user_created_idx ON forum_posts(user_id,created_at DESC);
CREATE TRIGGER forum_categories_set_updated_at BEFORE UPDATE ON forum_categories
 FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER forum_topics_set_updated_at BEFORE UPDATE ON forum_topics
 FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
CREATE TRIGGER forum_posts_set_updated_at BEFORE UPDATE ON forum_posts
 FOR EACH ROW EXECUTE FUNCTION set_row_updated_at();
ALTER TABLE moderation_actions
 ADD COLUMN forum_topic_id BIGINT REFERENCES forum_topics(id),
 ADD COLUMN forum_post_id BIGINT REFERENCES forum_posts(id);
ALTER TABLE moderation_actions DROP CONSTRAINT moderation_actions_action_check;
ALTER TABLE moderation_actions ADD CONSTRAINT moderation_actions_action_check CHECK(action IN
 ('ban','revoke','delete_message','role_change','hide_topic','restore_topic','close_topic','open_topic','pin_topic','unpin_topic','hide_post','restore_post'));
