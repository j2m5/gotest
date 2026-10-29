DELETE FROM moderation_actions WHERE action IN
 ('hide_topic','restore_topic','close_topic','open_topic','pin_topic','unpin_topic','hide_post','restore_post');
ALTER TABLE moderation_actions DROP COLUMN forum_post_id, DROP COLUMN forum_topic_id;
ALTER TABLE moderation_actions DROP CONSTRAINT moderation_actions_action_check;
ALTER TABLE moderation_actions ADD CONSTRAINT moderation_actions_action_check CHECK(action IN ('ban','revoke','delete_message','role_change'));
DROP TABLE forum_posts;
DROP TABLE forum_topics;
DROP TABLE forum_categories;
