package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/ability"
	"gotest/internal/character"
)

type AbilityRepository struct{ pool *pgxpool.Pool }

func NewAbilityRepository(pool *pgxpool.Pool) *AbilityRepository {
	return &AbilityRepository{pool: pool}
}

const learnedAbilitiesQuery = `SELECT a.id,a.code,a.name,COALESCE(a.description,''),a.icon,
 a.damage,a.heal,a.mp_cost,a.cooldown_ms,ca.ready_at FROM character_abilities ca
 JOIN characters c ON c.id=ca.character_id
 JOIN class_abilities allowed ON allowed.class_id=c.class_id AND allowed.ability_id=ca.ability_id
 JOIN abilities a ON a.id=ca.ability_id WHERE ca.character_id=$1 `

func scanOwnedAbility(row pgx.Row) (ability.Owned, error) {
	var owned ability.Owned
	err := row.Scan(&owned.ID, &owned.Code, &owned.Name, &owned.Description, &owned.Icon,
		&owned.Damage, &owned.Heal, &owned.MPCost, &owned.CooldownMS, &owned.ReadyAt)
	return owned, err
}

func (r *AbilityRepository) List(ctx context.Context, actor character.Actor, characterID int) (*ability.Loadout, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	if err = requireSelectedCharacter(ctx, tx, actor, characterID); err != nil {
		return nil, err
	}
	c, err := scanCharacter(tx.QueryRow(ctx, `SELECT `+characterColumns+` FROM characters WHERE id=$1 AND user_id=$2`, characterID, actor.UserID))
	if err != nil {
		return nil, err
	}
	result := &ability.Loadout{Character: c, ObservedAt: c.Regeneration.ObservedAt, GlobalCooldownMS: ability.GlobalCooldownDuration.Milliseconds(), Abilities: []ability.Owned{}}
	if err = tx.QueryRow(ctx, `SELECT global_cooldown_ready_at FROM characters WHERE id=$1`, characterID).Scan(&result.GlobalReadyAt); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, learnedAbilitiesQuery+` ORDER BY a.id`, characterID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		owned, err := scanOwnedAbility(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		owned.Availability = ability.AvailabilityAt(owned.Definition, owned.ReadyAt, result.GlobalReadyAt, c.HPCurrent, c.MPCurrent, result.ObservedAt)
		result.Abilities = append(result.Abilities, owned)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
