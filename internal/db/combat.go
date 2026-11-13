package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/character"
	"gotest/internal/combat"
)

type CombatRepository struct {
	pool   *pgxpool.Pool
	notify func(int)
}

func NewCombatRepository(pool *pgxpool.Pool) *CombatRepository { return &CombatRepository{pool: pool} }

// WithNotifications returns a configured copy; its nonblocking callback runs
// only after a committed change. The database is still the durable source.
func (r *CombatRepository) WithNotifications(notify func(int)) *CombatRepository {
	copy := *r
	copy.notify = notify
	return &copy
}

func (r *CombatRepository) notifyCharacter(characterID int) {
	if r.notify != nil {
		r.notify(characterID)
	}
}

const battleColumns = `id,character_id,location_id,status,npc_spawn_id,npc_template_id,
 npc_name,npc_level,npc_rank,npc_hp,npc_hp_current,created_at,ended_at`

func scanBattle(row pgx.Row) (*combat.Battle, error) {
	b := &combat.Battle{}
	err := row.Scan(&b.ID, &b.CharacterID, &b.LocationID, &b.Status, &b.NPC.SpawnID, &b.NPC.TemplateID,
		&b.NPC.Name, &b.NPC.Level, &b.NPC.Rank, &b.NPC.HP, &b.NPC.HPCurrent, &b.CreatedAt, &b.EndedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, combat.ErrNotFound
	}
	return b, err
}

func characterInBattle(ctx context.Context, tx pgx.Tx, characterID int) (bool, error) {
	var active bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM battles WHERE character_id=$1 AND status='active')`, characterID).Scan(&active)
	return active, err
}

func (r *CombatRepository) Current(ctx context.Context, actor character.Actor, expected int) (*combat.State, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	if err = requireSelectedCharacter(ctx, tx, actor, expected); err != nil {
		return nil, err
	}
	c, err := scanCharacter(tx.QueryRow(ctx, `SELECT `+characterColumns+` FROM characters WHERE id=$1 AND user_id=$2`, expected, actor.UserID))
	if err != nil {
		return nil, err
	}
	b, err := scanBattle(tx.QueryRow(ctx, `SELECT `+battleColumns+` FROM battles WHERE character_id=$1 AND status='active'`, expected))
	if err != nil && !errors.Is(err, combat.ErrNotFound) {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &combat.State{Character: c, Battle: b}, nil
}
