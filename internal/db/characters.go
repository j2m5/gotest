package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/character"
)

// CharacterRepository persists character state and serializes account mutations
// with logout through the users row. It never trusts session middleware alone.
type CharacterRepository struct{ pool *pgxpool.Pool }

func NewCharacterRepository(pool *pgxpool.Pool) *CharacterRepository {
	return &CharacterRepository{pool: pool}
}

func rollbackCharacterTx(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}

func (r *CharacterRepository) Catalog(ctx context.Context) (*character.Catalog, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	catalog := &character.Catalog{Factions: []character.Faction{}, Classes: []character.Class{}}
	// The catalog exposes display labels; stable codes identify the entries.
	rows, err := tx.Query(ctx, `SELECT id,code,COALESCE(NULLIF(BTRIM(alias),''),name),description,icon FROM factions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f character.Faction
		if err := rows.Scan(&f.ID, &f.Code, &f.Name, &f.Description, &f.Icon); err != nil {
			rows.Close()
			return nil, err
		}
		catalog.Factions = append(catalog.Factions, f)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, `SELECT id,faction_id,code,COALESCE(NULLIF(BTRIM(alias),''),name),description,icon FROM classes ORDER BY faction_id,id`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c character.Class
		if err := rows.Scan(&c.ID, &c.FactionID, &c.Code, &c.Name, &c.Description, &c.Icon); err != nil {
			rows.Close()
			return nil, err
		}
		catalog.Classes = append(catalog.Classes, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return catalog, nil
}

func authenticateCharacterActor(ctx context.Context, tx pgx.Tx, actor character.Actor, lock bool) error {
	if lock {
		var id int
		err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, actor.UserID).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return character.ErrUnauthenticated
		}
		if err != nil {
			return err
		}
	}
	// Separate statement after acquiring the account lock obtains a fresh READ
	// COMMITTED snapshot, including any logout that committed while we waited.
	var authenticated bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users u JOIN sessions s ON s.user_id=u.id WHERE u.id=$1 AND s.token=$2 AND s.expires_at>clock_timestamp() AND u.email_verified_at IS NOT NULL)`, actor.UserID, actor.SessionHash).Scan(&authenticated)
	if err != nil {
		return err
	}
	if !authenticated {
		return character.ErrUnauthenticated
	}
	return nil
}

const characterColumns = `id,name,faction_id,class_id,role,level,experience,avatar,hp,mp,hp_current,mp_current`

func scanCharacter(row pgx.Row) (*character.Character, error) {
	c := &character.Character{}
	err := row.Scan(&c.ID, &c.Name, &c.FactionID, &c.ClassID, &c.Role, &c.Level, &c.Experience, &c.Avatar, &c.HP, &c.MP, &c.HPCurrent, &c.MPCurrent)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CharacterRepository) List(ctx context.Context, actor character.Actor) (*character.Roster, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err := authenticateCharacterActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	roster := &character.Roster{Characters: []character.Character{}}
	if err := tx.QueryRow(ctx, `SELECT character_id FROM users WHERE id=$1`, actor.UserID).Scan(&roster.ActiveCharacterID); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT `+characterColumns+` FROM characters WHERE user_id=$1 ORDER BY id`, actor.UserID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		c, err := scanCharacter(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		roster.Characters = append(roster.Characters, *c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return roster, nil
}

func characterStorageError(err error) error {
	var postgres *pgconn.PgError
	if errors.As(err, &postgres) {
		if postgres.Code == "23505" && (postgres.ConstraintName == "characters_name_normalized_key" || postgres.ConstraintName == "characters_name_key") {
			return character.ErrNameTaken
		}
		if postgres.Code == "23503" && postgres.ConstraintName == "characters_class_faction_fkey" {
			return character.ErrInvalidClass
		}
	}
	return err
}

func (r *CharacterRepository) Create(ctx context.Context, actor character.Actor, in character.CreateInput) (*character.Character, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err := authenticateCharacterActor(ctx, tx, actor, true); err != nil {
		return nil, err
	}
	// INSERT SELECT validates both catalog identifiers together; defaults belong
	// to SQL so future balance changes do not require a duplicate Go constant.
	c, err := scanCharacter(tx.QueryRow(ctx, `INSERT INTO characters(user_id,name,faction_id,class_id) SELECT $1,$2,faction_id,id FROM classes WHERE faction_id=$3 AND id=$4 RETURNING `+characterColumns, actor.UserID, in.Name, in.FactionID, in.ClassID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, character.ErrInvalidClass
	}
	if err != nil {
		return nil, characterStorageError(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO character_abilities(character_id,ability_id) SELECT $1,ability_id FROM class_abilities WHERE class_id=$2`, c.ID, c.ClassID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, characterStorageError(err)
	}
	return c, nil
}

func (r *CharacterRepository) SetActive(ctx context.Context, actor character.Actor, id *int) (*character.Character, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err := authenticateCharacterActor(ctx, tx, actor, true); err != nil {
		return nil, err
	}
	var c *character.Character
	if id != nil {
		c, err = scanCharacter(tx.QueryRow(ctx, `SELECT `+characterColumns+` FROM characters WHERE id=$1 AND user_id=$2 FOR KEY SHARE`, *id, actor.UserID))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, character.ErrNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET character_id=$2 WHERE id=$1`, actor.UserID, id); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c, nil
}
