package db

import (
	"context"
	"errors"
	"gotest/internal/auth"
	"gotest/internal/forum"
	"gotest/internal/social"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Forum uses the same account locks and session policy as chat and moderation.
type ForumRepository struct{ shared *SocialRepository }

var _ forum.Repository = (*ForumRepository)(nil)

func NewForumRepository(p *pgxpool.Pool) *ForumRepository {
	return &ForumRepository{NewSocialRepository(p)}
}
func forumStaff(role int16) bool { return role == auth.RoleModerator || role == auth.RoleAdministrator }
func forumNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return forum.ErrNotFound
	}
	return err
}

const forumCategoryColumns = `c.id,c.code,c.title,c.description,role.name`

func scanForumCategory(row pgx.Row) (forum.Category, error) {
	var c forum.Category
	e := row.Scan(&c.ID, &c.Code, &c.Title, &c.Description, &c.CreateRole)
	return c, e
}

const forumTopicSummaryColumns = `t.id,t.category_id,t.user_id,t.character_id,t.author_name,role.name,t.title,t.created_at,t.updated_at,t.closed_at,t.pinned_at,t.hidden_at,(SELECT count(*) FROM forum_posts p WHERE p.topic_id=t.id)`
const forumTopicColumns = forumTopicSummaryColumns + `,t.body,t.edited_at,t.version`
const forumTopicFrom = ` FROM forum_topics t JOIN users u ON u.id=t.user_id JOIN account_roles role ON role.id=u.role_id `

func scanForumTopic(row pgx.Row) (*forum.Topic, error) {
	v := &forum.Topic{}
	e := row.Scan(&v.ID, &v.CategoryID, &v.UserID, &v.CharacterID, &v.AuthorName, &v.AuthorRole, &v.Title, &v.CreatedAt, &v.UpdatedAt, &v.ClosedAt, &v.PinnedAt, &v.HiddenAt, &v.ReplyCount, &v.Body, &v.EditedAt, &v.Version)
	return v, forumNotFound(e)
}
func loadForumTopic(ctx context.Context, tx pgx.Tx, id int64, lock bool) (*forum.Topic, error) {
	suffix := ""
	if lock {
		suffix = " FOR UPDATE OF t"
	}
	return scanForumTopic(tx.QueryRow(ctx, `SELECT `+forumTopicColumns+forumTopicFrom+`WHERE t.id=$1`+suffix, id))
}

const forumPostColumns = `p.id,p.topic_id,p.user_id,p.character_id,p.author_name,role.name,p.body,p.created_at,p.updated_at,p.edited_at,p.hidden_at,p.version`
const forumPostFrom = ` FROM forum_posts p JOIN users u ON u.id=p.user_id JOIN account_roles role ON role.id=u.role_id `

func scanForumPost(row pgx.Row) (*forum.Post, error) {
	v := &forum.Post{}
	e := row.Scan(&v.ID, &v.TopicID, &v.UserID, &v.CharacterID, &v.AuthorName, &v.AuthorRole, &v.Body, &v.CreatedAt, &v.UpdatedAt, &v.EditedAt, &v.HiddenAt, &v.Version)
	return v, forumNotFound(e)
}
func loadForumPost(ctx context.Context, tx pgx.Tx, id int64, lock bool) (*forum.Post, error) {
	suffix := ""
	if lock {
		suffix = " FOR UPDATE OF p"
	}
	return scanForumPost(tx.QueryRow(ctx, `SELECT `+forumPostColumns+forumPostFrom+`WHERE p.id=$1`+suffix, id))
}

func (r *ForumRepository) Categories(ctx context.Context, a forum.Actor) (*forum.CategoryList, error) {
	tx, _, e := r.shared.begin(ctx, a, 0)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	out := &forum.CategoryList{Categories: []forum.Category{}}
	rows, e := tx.Query(ctx, `SELECT `+forumCategoryColumns+` FROM forum_categories c JOIN account_roles role ON role.id=c.create_role_id ORDER BY c.position,c.id`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanForumCategory(rows)
		if err != nil {
			return nil, err
		}
		out.Categories = append(out.Categories, c)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (r *ForumRepository) Topics(ctx context.Context, a forum.Actor, in forum.TopicsInput) (*forum.TopicList, error) {
	pinned, id, e := forum.ParseCursor(in.Cursor)
	if e != nil {
		return nil, e
	}
	tx, users, e := r.shared.begin(ctx, a, 0)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	if in.IncludeHidden && !forumStaff(users[a.UserID].role) {
		return nil, social.ErrForbidden
	}
	category, e := scanForumCategory(tx.QueryRow(ctx, `SELECT `+forumCategoryColumns+` FROM forum_categories c JOIN account_roles role ON role.id=c.create_role_id WHERE c.id=$1`, in.CategoryID))
	if e != nil {
		return nil, forumNotFound(e)
	}
	out := &forum.TopicList{Category: category, Topics: []forum.TopicSummary{}}
	out.ForumBan, e = activeBan(ctx, tx, a.UserID, "forum")
	if e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT `+forumTopicColumns+forumTopicFrom+`WHERE t.category_id=$1 AND ($2 OR t.hidden_at IS NULL) AND ($3::bigint=0 OR ((t.pinned_at IS NOT NULL),t.id)<($4::boolean,$3::bigint)) ORDER BY (t.pinned_at IS NOT NULL) DESC,t.id DESC LIMIT 31`, in.CategoryID, in.IncludeHidden, id, pinned)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanForumTopic(rows)
		if err != nil {
			return nil, err
		}
		out.Topics = append(out.Topics, v.TopicSummary)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if len(out.Topics) > 30 {
		out.Topics = out.Topics[:30]
		last := out.Topics[29]
		prefix := "n:"
		if last.PinnedAt != nil {
			prefix = "p:"
		}
		cursor := prefix + strconv.FormatInt(last.ID, 10)
		out.NextCursor = &cursor
	}
	return out, tx.Commit(ctx)
}
func (r *ForumRepository) Topic(ctx context.Context, a forum.Actor, id, after int64) (*forum.TopicPage, error) {
	tx, users, e := r.shared.begin(ctx, a, 0)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	topic, e := loadForumTopic(ctx, tx, id, false)
	if e != nil {
		return nil, e
	}
	staff := forumStaff(users[a.UserID].role)
	if topic.HiddenAt != nil && !staff {
		return nil, forum.ErrNotFound
	}
	out := &forum.TopicPage{Topic: *topic, Posts: []forum.Post{}}
	out.ForumBan, e = activeBan(ctx, tx, a.UserID, "forum")
	if e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT `+forumPostColumns+forumPostFrom+`WHERE p.topic_id=$1 AND p.id>$2 ORDER BY p.id LIMIT 31`, id, after)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanForumPost(rows)
		if err != nil {
			return nil, err
		}
		if v.HiddenAt != nil && !staff {
			v.Body = nil
		}
		out.Posts = append(out.Posts, *v)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if len(out.Posts) > 30 {
		out.Posts = out.Posts[:30]
		last := out.Posts[29].ID
		out.NextAfterID = &last
	}
	return out, tx.Commit(ctx)
}

func forumCharacter(ctx context.Context, tx pgx.Tx, a forum.Actor, u socialUser, expected int) (string, error) {
	b, e := activeBan(ctx, tx, a.UserID, "forum")
	if e != nil {
		return "", e
	}
	if b != nil {
		return "", forum.ErrBanned
	}
	if u.character == nil {
		return "", social.ErrNoCharacter
	}
	if *u.character != expected {
		return "", social.ErrCharacterChanged
	}
	var name string
	e = tx.QueryRow(ctx, `SELECT name FROM characters WHERE id=$1 AND user_id=$2`, expected, a.UserID).Scan(&name)
	if errors.Is(e, pgx.ErrNoRows) {
		return "", social.ErrNoCharacter
	}
	return name, e
}
func forumCooldown(ctx context.Context, tx pgx.Tx, user int) error {
	var fast bool
	e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM forum_topics WHERE user_id=$1 AND created_at>clock_timestamp()-interval '5 seconds') OR EXISTS(SELECT 1 FROM forum_posts WHERE user_id=$1 AND created_at>clock_timestamp()-interval '5 seconds')`, user).Scan(&fast)
	if e != nil {
		return e
	}
	if fast {
		return forum.ErrTooFast
	}
	return nil
}
func forumWritableTopic(v *forum.Topic, role int16) error {
	if v.HiddenAt != nil {
		if !forumStaff(role) {
			return forum.ErrNotFound
		}
		return forum.ErrClosed
	}
	if v.ClosedAt != nil {
		return forum.ErrClosed
	}
	return nil
}

func (r *ForumRepository) Create(ctx context.Context, a forum.Actor, in forum.CreateInput) (*forum.Topic, error) {
	tx, users, e := r.shared.begin(ctx, a, 0)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	name, e := forumCharacter(ctx, tx, a, users[a.UserID], in.CharacterID)
	if e != nil {
		return nil, e
	}
	var role int16
	e = tx.QueryRow(ctx, `SELECT create_role_id FROM forum_categories WHERE id=$1`, in.CategoryID).Scan(&role)
	if e != nil {
		return nil, forumNotFound(e)
	}
	if users[a.UserID].role < role {
		return nil, social.ErrForbidden
	}
	if e = forumCooldown(ctx, tx, a.UserID); e != nil {
		return nil, e
	}
	var id int64
	e = tx.QueryRow(ctx, `INSERT INTO forum_topics(category_id,user_id,character_id,author_name,title,body) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, in.CategoryID, a.UserID, in.CharacterID, name, in.Title, in.Body).Scan(&id)
	if e != nil {
		return nil, e
	}
	out, e := loadForumTopic(ctx, tx, id, false)
	if e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (r *ForumRepository) Reply(ctx context.Context, a forum.Actor, in forum.ReplyInput) (*forum.Post, error) {
	tx, users, e := r.shared.begin(ctx, a, 0)
	if e != nil {
		return nil, e
	}
	defer rollbackSocialTx(tx)
	topic, e := loadForumTopic(ctx, tx, in.TopicID, true)
	if e != nil {
		return nil, e
	}
	if e = checkSocialSession(ctx, tx, a); e != nil {
		return nil, e
	}
	if e = forumWritableTopic(topic, users[a.UserID].role); e != nil {
		return nil, e
	}
	name, e := forumCharacter(ctx, tx, a, users[a.UserID], in.CharacterID)
	if e != nil {
		return nil, e
	}
	if e = forumCooldown(ctx, tx, a.UserID); e != nil {
		return nil, e
	}
	var id int64
	e = tx.QueryRow(ctx, `INSERT INTO forum_posts(topic_id,user_id,character_id,author_name,body) VALUES($1,$2,$3,$4,$5) RETURNING id`, in.TopicID, a.UserID, in.CharacterID, name, in.Body).Scan(&id)
	if e != nil {
		return nil, e
	}
	out, e := loadForumPost(ctx, tx, id, false)
	if e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
