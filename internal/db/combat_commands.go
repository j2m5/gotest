package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/ability"
	"gotest/internal/character"
	"gotest/internal/combat"
)

func lockPlayerExecution(ctx context.Context, tx pgx.Tx, actor character.Actor, battleID int64, characterID int) (combat.Frame, error) {
	if err := authenticateGameActor(ctx, tx, actor, true); err != nil {
		return combat.Frame{}, err
	}
	if _, _, err := worldCharacter(ctx, tx, actor, characterID, true); err != nil {
		return combat.Frame{}, err
	}
	b, err := lockCombatBattle(ctx, tx, characterID, battleID)
	if err != nil {
		return combat.Frame{}, err
	}
	f, err := loadExecution(ctx, tx, b)
	if err != nil {
		return f, err
	}
	return f, authenticateGameActor(ctx, tx, actor, false)
}

func (r *CombatRepository) Use(ctx context.Context, actor character.Actor, battleID int64, in combat.UseInput) (*combat.Result, error) {
	if err := combat.ValidateUse(battleID, in); err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	before, err := lockPlayerExecution(ctx, tx, actor, battleID, in.CharacterID)
	if err != nil {
		return nil, err
	}
	var recordedAbility, sequence int
	err = tx.QueryRow(ctx, `SELECT ability_id,event_sequence FROM battle_commands WHERE battle_id=$1 AND command_key=$2`, battleID, in.Key).Scan(&recordedAbility, &sequence)
	replayed := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	var owned ability.Owned
	var actionErr error
	if !replayed && before.Battle.Status == combat.Active {
		owned, actionErr = lockCombatAbility(ctx, tx, in.CharacterID, in.AbilityID)
		if actionErr != nil && !errors.Is(actionErr, ability.ErrUnavailable) {
			return nil, actionErr
		}
	}
	// A queued command cannot use an expired or revoked session after acquiring
	// the ability lock. All subsequent effects share this authoritative instant.
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	var at time.Time
	var global *time.Time
	var power, mastery int
	err = tx.QueryRow(ctx, `SELECT clock_timestamp(),global_cooldown_ready_at,power,mastery FROM characters WHERE id=$1`, in.CharacterID).Scan(&at, &global, &power, &mastery)
	if err != nil {
		return nil, err
	}
	f, events, err := combat.Advance(before, at)
	if err != nil {
		return nil, err
	}
	if replayed {
		if recordedAbility != in.AbilityID {
			actionErr = combat.ErrConflict
		}
	} else if f.Battle.Status != combat.Active {
		actionErr = combat.ErrFinished
	} else if actionErr == nil {
		var use *ability.Use
		var resources character.Resources
		use, resources, actionErr = prepareOwnedUse(ctx, tx, in.CharacterID, f.Resources, owned, global, at)
		if actionErr != nil && !readinessFailure(actionErr) {
			return nil, actionErr
		}
		if actionErr == nil {
			f.Resources = resources
			var event combat.Event
			f, event, err = combat.PlayerEffect(f, use.Ability.Damage, use.Ability.Heal, power, mastery, use.Ability.Code, use.Ability.Name, at)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
			sequence = event.Sequence
		}
	}
	if err = r.saveExecution(ctx, tx, before, &f, events); err != nil {
		return nil, err
	}
	if !replayed && actionErr == nil {
		_, err = tx.Exec(ctx, `INSERT INTO battle_commands(battle_id,command_key,ability_id,event_sequence) VALUES($1,$2,$3,$4)`, battleID, in.Key, in.AbilityID, sequence)
		if err != nil {
			return nil, err
		}
	}
	return r.commitCombatResult(ctx, tx, &f.Battle, replayed, sequence, actionErr, len(events) > 0)
}

func readinessFailure(err error) bool {
	return errors.Is(err, ability.ErrUnavailable) || errors.Is(err, ability.ErrCooldown) || errors.Is(err, ability.ErrGlobalCooldown) || errors.Is(err, ability.ErrDefeated) || errors.Is(err, character.ErrInsufficientEnergy)
}

func (r *CombatRepository) Retreat(ctx context.Context, actor character.Actor, battleID int64, characterID int) (*combat.Result, error) {
	if battleID <= 0 || !combat.ValidID(characterID) {
		return nil, combat.ErrInvalidInput
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	before, err := lockPlayerExecution(ctx, tx, actor, battleID, characterID)
	if err != nil {
		return nil, err
	}
	at, err := resourceClock(ctx, tx)
	if err != nil {
		return nil, err
	}
	f, events, err := combat.Advance(before, at)
	if err != nil {
		return nil, err
	}
	replayed := f.Battle.Status != combat.Active
	if !replayed {
		f.Finish(combat.Retreated, at)
		events = append(events, f.Record(combat.Event{Kind: "retreated", Actor: "player", OccurredAt: at}))
	}
	if err = r.saveExecution(ctx, tx, before, &f, events); err != nil {
		return nil, err
	}
	return r.commitCombatResult(ctx, tx, &f.Battle, replayed, f.Sequence, nil, len(events) > 0)
}
