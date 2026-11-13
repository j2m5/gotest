package db

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/progression"
	"gotest/internal/wallet"
)

func loadExecution(ctx context.Context, tx pgx.Tx, b *combat.Battle) (combat.Frame, error) {
	f := combat.Frame{Battle: *b}
	var rules []byte
	var next *time.Time
	err := tx.QueryRow(ctx, `SELECT rules,pattern_step,next_action_at,expires_at,event_sequence FROM battle_executions WHERE battle_id=$1 FOR UPDATE`, b.ID).
		Scan(&rules, &f.PatternStep, &next, &f.ExpiresAt, &f.Sequence)
	if errors.Is(err, pgx.ErrNoRows) {
		return f, combat.ErrNotFound
	}
	if err != nil {
		return f, err
	}
	if err = json.Unmarshal(rules, &f.Rules); err != nil {
		return f, err
	}
	if err = f.Rules.Validate(); err != nil {
		return f, err
	}
	if next != nil {
		f.NextActionAt = *next
	}
	if f.Rules.HP != b.NPC.HP {
		return f, combat.ErrInvalidState
	}
	f.Resources, err = lockCharacterResources(ctx, tx, b.CharacterID)
	if err != nil {
		return f, err
	}
	err = tx.QueryRow(ctx, `SELECT defense FROM characters WHERE id=$1`, b.CharacterID).Scan(&f.Defense)
	return f, err
}

func (r *CombatRepository) Start(ctx context.Context, actor character.Actor, in combat.StartInput) (*combat.Result, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	change, err := r.startCombatInTx(ctx, tx, actor, in, combat.NPCStats{HP: 1}, true)
	if err != nil {
		return nil, err
	}
	b := change.Battle
	sequence := 0
	if !change.Replayed {
		rules, err := combat.RulesFor(b.NPC.Level, b.NPC.Rank)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(rules)
		if err != nil {
			return nil, err
		}
		resources, err := lockCharacterResources(ctx, tx, b.CharacterID)
		if err != nil {
			return nil, err
		}
		f := combat.Frame{Battle: *b, Rules: rules, Resources: resources,
			NextActionAt: b.CreatedAt.Add(time.Duration(rules.Pattern[0].DelayMS) * time.Millisecond), ExpiresAt: b.CreatedAt.Add(time.Duration(rules.DurationMS) * time.Millisecond)}
		event := f.Record(combat.Event{Kind: "started", Actor: "player", OccurredAt: b.CreatedAt})
		_, err = tx.Exec(ctx, `INSERT INTO battle_executions(battle_id,rules,next_action_at,expires_at,event_sequence,created_at)
 VALUES($1,$2,$3,$4,$5,$6)`, b.ID, encoded, f.NextActionAt, f.ExpiresAt, f.Sequence, b.CreatedAt)
		if err != nil {
			return nil, err
		}
		if err = appendCombatEvents(ctx, tx, b.ID, []combat.Event{event}); err != nil {
			return nil, err
		}
		sequence = event.Sequence
	} else {
		// Start always acknowledges the original first event, even if the
		// encounter has advanced since the request was first accepted.
		if err = tx.QueryRow(ctx, `SELECT 1 FROM battle_executions WHERE battle_id=$1`, b.ID).Scan(&sequence); err != nil {
			return nil, err
		}
	}
	return r.commitCombatResult(ctx, tx, b, change.Replayed, sequence, nil, !change.Replayed)
}

// saveExecution commits no transaction itself. Every effect, log entry, reward,
// NPC release and timer update belongs to the caller's single atomic operation.
func (r *CombatRepository) saveExecution(ctx context.Context, tx pgx.Tx, before combat.Frame, f *combat.Frame, events []combat.Event) error {
	if len(events) == 0 {
		return nil
	}
	if f.Battle.Status == combat.Won && before.Battle.Status == combat.Active {
		key := "pve:" + strconv.FormatInt(f.Battle.ID, 10)
		xp, err := NewProgressionRepository(r.pool).GrantInTx(ctx, tx, progression.Reward{CharacterID: f.Battle.CharacterID, Key: key, Experience: f.Rules.Experience})
		if err != nil {
			return err
		}
		_, err = NewWalletRepository(r.pool).ApplyInTx(ctx, tx, wallet.Operation{CharacterID: f.Battle.CharacterID, Currency: wallet.Credits, Amount: f.Rules.Credits, Source: "pve", Key: key, Description: "Победа над " + f.Battle.NPC.Name})
		if err != nil {
			return err
		}
		events = append(events, f.Record(combat.Event{Kind: "won", Actor: "system", Experience: xp.Change.Granted, Credits: f.Rules.Credits, OccurredAt: *f.Battle.EndedAt}))
	}
	if f.Resources != before.Resources {
		if err := saveCharacterResources(ctx, tx, f.Battle.CharacterID, f.Resources); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx, `UPDATE battles SET status=$2,npc_hp_current=$3,ended_at=$4 WHERE id=$1`, f.Battle.ID, f.Battle.Status, f.Battle.NPC.HPCurrent, f.Battle.EndedAt)
	if err != nil {
		return err
	}
	var next *time.Time
	if !f.NextActionAt.IsZero() {
		next = &f.NextActionAt
	}
	_, err = tx.Exec(ctx, `UPDATE battle_executions SET pattern_step=$2,next_action_at=$3,event_sequence=$4,
 retry_count=0,retry_at=NULL WHERE battle_id=$1`, f.Battle.ID, f.PatternStep, next, f.Sequence)
	if err != nil {
		return err
	}
	if f.Battle.Status != combat.Active {
		var respawn *time.Time
		if f.Battle.Status == combat.Won {
			at := f.Battle.EndedAt.Add(time.Duration(f.Rules.RespawnMS) * time.Millisecond)
			respawn = &at
		}
		if _, err = tx.Exec(ctx, `UPDATE npc_spawns SET respawn_at=$2 WHERE id=$1`, f.Battle.NPC.SpawnID, respawn); err != nil {
			return err
		}
	}
	return appendCombatEvents(ctx, tx, f.Battle.ID, events)
}

func (r *CombatRepository) commitCombatResult(ctx context.Context, tx pgx.Tx, b *combat.Battle, replayed bool, sequence int, actionErr error, changed bool) (*combat.Result, error) {
	c, err := scanCharacter(tx.QueryRow(ctx, `SELECT `+characterColumns+` FROM characters WHERE id=$1`, b.CharacterID))
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	if changed {
		r.notifyCharacter(b.CharacterID)
	}
	if actionErr != nil {
		return nil, actionErr
	}
	return &combat.Result{State: combat.State{Character: c, Battle: b}, Replayed: replayed, EventSequence: sequence}, nil
}
