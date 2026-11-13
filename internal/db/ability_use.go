package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/ability"
	"gotest/internal/character"
)

// PrepareUseInTx is a trusted combat building block, NOT a standalone cast.
// The caller authorizes the actor/target/battle and deduplicates the command
// before calling, then applies its effect and event in THIS transaction.
// Roll back on ANY error. Lock order: character, learned ability. CD/GCD survive
// session and battle changes because they belong to the character, not a battle.
func (r *AbilityRepository) PrepareUseInTx(ctx context.Context, tx pgx.Tx, characterID, abilityID int) (*ability.Use, error) {
	if characterID <= 0 || int64(characterID) > 1<<31-1 || abilityID <= 0 || int64(abilityID) > 1<<31-1 {
		return nil, ability.ErrInvalidInput
	}
	resources, err := lockCharacterResources(ctx, tx, characterID)
	if err != nil {
		return nil, err
	}
	owned, err := lockCombatAbility(ctx, tx, characterID, abilityID)
	if err != nil {
		return nil, err
	}
	var at time.Time
	var global *time.Time
	// Read the clock after both locks: lock contention cannot make a stale time
	// incorrectly approve a cast or shorten its newly recorded cooldown.
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp(),global_cooldown_ready_at FROM characters WHERE id=$1`, characterID).Scan(&at, &global); err != nil {
		return nil, err
	}
	use, _, err := prepareOwnedUse(ctx, tx, characterID, resources, owned, global, at)
	return use, err
}

func lockCombatAbility(ctx context.Context, tx pgx.Tx, characterID, abilityID int) (ability.Owned, error) {
	owned, err := scanOwnedAbility(tx.QueryRow(ctx, learnedAbilitiesQuery+` AND a.id=$2 FOR UPDATE OF ca FOR SHARE OF a`, characterID, abilityID))
	if errors.Is(err, pgx.ErrNoRows) {
		err = ability.ErrUnavailable
	}
	return owned, err
}

// The executor has already locked the character and learned ability and resolved
// all NPC actions due at this database timestamp. Domain readiness errors occur
// before any writes; SQL errors still require rolling back the whole transaction.
func prepareOwnedUse(ctx context.Context, tx pgx.Tx, characterID int, resources character.Resources, owned ability.Owned, global *time.Time, at time.Time) (*ability.Use, character.Resources, error) {
	hp, _ := character.RecoverResourceWithRemainder(resources.Health.Current, resources.Health.Maximum, resources.Health.Remainder, resources.Health.UpdatedAt, at, character.HealthRecoveryPeriod)
	mp, _ := character.RecoverResourceWithRemainder(resources.Energy.Current, resources.Energy.Maximum, resources.Energy.Remainder, resources.Energy.UpdatedAt, at, character.EnergyRecoveryPeriod)
	if err := ability.ReadinessError(ability.AvailabilityAt(owned.Definition, owned.ReadyAt, global, hp, mp, at)); err != nil {
		return nil, resources, err
	}
	next, err := character.ChangeResources(resources, character.ResourceChange{MP: -owned.MPCost}, at)
	if err != nil {
		return nil, resources, err
	}
	if next != resources {
		if err = saveCharacterResources(ctx, tx, characterID, next); err != nil {
			return nil, resources, err
		}
	}
	result := &ability.Use{Ability: owned.Definition, UsedAt: at, ReadyAt: at.Add(time.Duration(owned.CooldownMS) * time.Millisecond), GlobalReadyAt: at.Add(ability.GlobalCooldownDuration), EnergyRemaining: mp - owned.MPCost}
	if _, err = tx.Exec(ctx, `UPDATE characters SET global_cooldown_ready_at=$2 WHERE id=$1`, characterID, result.GlobalReadyAt); err != nil {
		return nil, resources, err
	}
	if _, err = tx.Exec(ctx, `UPDATE character_abilities SET ready_at=$3 WHERE character_id=$1 AND ability_id=$2`, characterID, owned.ID, result.ReadyAt); err != nil {
		return nil, resources, err
	}
	return result, next, nil
}
