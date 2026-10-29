package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/character"
	"gotest/internal/presence"
)

type PresenceRepository struct{ pool *pgxpool.Pool }

func NewPresenceRepository(pool *pgxpool.Pool) *PresenceRepository {
	return &PresenceRepository{pool: pool}
}

// This predicate is shared by count, batch status and directory. CURRENT_TIMESTAMP
// is stable for the whole read transaction, including the expiry boundary.
const onlineCharacters = `WITH online AS (
	SELECT DISTINCT u.character_id AS id
	FROM sessions s JOIN users u ON u.id=s.user_id
	WHERE u.character_id=s.presence_character_id AND u.email_verified_at IS NOT NULL
	AND s.expires_at>CURRENT_TIMESTAMP
	AND s.last_seen_at>CURRENT_TIMESTAMP-($1 * interval '1 second')
	AND NOT EXISTS(SELECT 1 FROM user_bans b WHERE b.user_id=u.id AND b.scope='account'
		AND b.revoked_at IS NULL AND (b.expires_at IS NULL OR b.expires_at>CURRENT_TIMESTAMP))
) `

func authenticatePresenceActor(ctx context.Context, tx pgx.Tx, actor character.Actor, lock bool) error {
	if err := authenticateCharacterActor(ctx, tx, actor, lock); err != nil {
		return err
	}
	var banned bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_bans WHERE user_id=$1 AND scope='account'
		AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at>clock_timestamp()))`, actor.UserID).Scan(&banned)
	if err != nil {
		return err
	}
	if banned {
		return character.ErrUnauthenticated
	}
	return nil
}

// Call after acquiring the account lock and checking the session/selection.
// force is used for an explicit selection, which is itself a player action.
func recordCharacterActivity(ctx context.Context, tx pgx.Tx, actor character.Actor, id *int, force bool) error {
	_, err := tx.Exec(ctx, `UPDATE sessions SET presence_character_id=$3,
		last_seen_at=CASE WHEN $3::integer IS NULL THEN NULL ELSE clock_timestamp() END
		WHERE user_id=$1 AND token=$2 AND ($4 OR presence_character_id IS DISTINCT FROM $3
		OR ($3::integer IS NOT NULL AND (last_seen_at IS NULL OR last_seen_at<=clock_timestamp()-($5 * interval '1 second')))
		OR ($3::integer IS NULL AND last_seen_at IS NOT NULL))`, actor.UserID, actor.SessionHash, id, force, presence.WriteInterval.Seconds())
	return err
}

func (r *PresenceRepository) Heartbeat(ctx context.Context, actor character.Actor, expected *int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackCharacterTx(tx)
	if err := authenticatePresenceActor(ctx, tx, actor, true); err != nil {
		return err
	}
	var current *int
	if err := tx.QueryRow(ctx, `SELECT character_id FROM users WHERE id=$1`, actor.UserID).Scan(&current); err != nil {
		return err
	}
	if (current == nil) != (expected == nil) || (current != nil && *current != *expected) {
		return presence.ErrCharacterChanged
	}
	if err := recordCharacterActivity(ctx, tx, actor, current, false); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PresenceRepository) read(ctx context.Context, actor character.Actor) (pgx.Tx, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	if err := authenticatePresenceActor(ctx, tx, actor, false); err != nil {
		rollbackCharacterTx(tx)
		return nil, err
	}
	return tx, nil
}

func (r *PresenceRepository) Snapshot(ctx context.Context, actor character.Actor, ids []int) (*presence.Snapshot, error) {
	tx, err := r.read(ctx, actor)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	result := &presence.Snapshot{Characters: make([]presence.Status, 0, len(ids))}
	if err := tx.QueryRow(ctx, onlineCharacters+`SELECT count(*) FROM online`, presence.OnlineWindow.Seconds()).Scan(&result.Total); err != nil {
		return nil, err
	}
	online := make(map[int]bool, len(ids))
	if len(ids) > 0 {
		rows, err := tx.Query(ctx, onlineCharacters+`SELECT id FROM online WHERE id=ANY($2::integer[])`, presence.OnlineWindow.Seconds(), ids)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			online[id] = true
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	for _, id := range ids {
		result.Characters = append(result.Characters, presence.Status{CharacterID: id, IsOnline: online[id]})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PresenceRepository) Online(ctx context.Context, actor character.Actor, after int) (*presence.Directory, error) {
	tx, err := r.read(ctx, actor)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	result := &presence.Directory{Characters: []character.PublicProfile{}}
	if err := tx.QueryRow(ctx, onlineCharacters+`SELECT count(*) FROM online`, presence.OnlineWindow.Seconds()).Scan(&result.Total); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, onlineCharacters+`SELECT c.id,c.name,c.faction_id,c.class_id,c.level,c.avatar,role.name
		FROM online JOIN characters c ON c.id=online.id JOIN users u ON u.id=c.user_id
		JOIN account_roles role ON role.id=u.role_id WHERE c.id>$2 ORDER BY c.id LIMIT $3`, presence.OnlineWindow.Seconds(), after, presence.PageSize+1)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c character.PublicProfile
		if err := rows.Scan(&c.ID, &c.Name, &c.FactionID, &c.ClassID, &c.Level, &c.Avatar, &c.AccountRole); err != nil {
			rows.Close()
			return nil, err
		}
		result.Characters = append(result.Characters, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result.Characters) > presence.PageSize {
		result.Characters = result.Characters[:presence.PageSize]
		result.NextAfterID = new(result.Characters[presence.PageSize-1].ID)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
