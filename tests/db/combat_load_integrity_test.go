package db_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/character"
	"gotest/internal/combat"
)

type loadResourceAudit struct {
	stored, replayed                      character.Resources
	npcInitial, npcStored, npc            int
	started, healthChanged, energyChanged bool
}

// The load fixture starts with full resources and uses a damaging starter.
// Replay the committed ledger through the independently tested resource rules:
// this verifies persistence under concurrency, not just legal column ranges.
func checkCombatLoadResources(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `SELECT b.id,b.npc_hp,b.npc_hp_current,
 c.hp,c.hp_current,c.hp_updated_at,c.hp_regen_remainder,
 c.mp,c.mp_current,c.mp_updated_at,c.mp_regen_remainder
 FROM battles b JOIN characters c ON c.id=b.character_id`)
	if err != nil {
		return err
	}
	audits := make(map[int64]*loadResourceAudit)
	for rows.Next() {
		var id int64
		a := &loadResourceAudit{}
		hp, mp := &a.stored.Health, &a.stored.Energy
		if err := rows.Scan(&id, &a.npcInitial, &a.npcStored,
			&hp.Maximum, &hp.Current, &hp.UpdatedAt, &hp.Remainder,
			&mp.Maximum, &mp.Current, &mp.UpdatedAt, &mp.Remainder); err != nil {
			rows.Close()
			return err
		}
		a.npc = a.npcInitial
		audits[id] = a
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	rows, err = pool.Query(ctx, `SELECT e.battle_id,e.payload,a.mp_cost,a.heal
 FROM battle_events e LEFT JOIN battle_commands c ON c.battle_id=e.battle_id AND c.event_sequence=e.sequence
 LEFT JOIN abilities a ON a.id=c.ability_id ORDER BY e.battle_id,e.sequence`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var payload []byte
		var cost, heal *int
		if err := rows.Scan(&id, &payload, &cost, &heal); err != nil {
			return err
		}
		a := audits[id]
		if a == nil {
			return fmt.Errorf("load event refers to missing battle %d", id)
		}
		var event combat.Event
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}
		change := character.ResourceChange{}
		switch event.Kind {
		case "started":
			if a.started || event.Sequence != 1 {
				return fmt.Errorf("invalid load start event for %d", id)
			}
			a.started = true
			a.replayed = character.Resources{
				Health: character.Resource{Maximum: event.CharacterHP, Current: event.CharacterHP, UpdatedAt: event.OccurredAt},
				Energy: character.Resource{Maximum: event.CharacterMP, Current: event.CharacterMP, UpdatedAt: event.OccurredAt},
			}
		case "npc_ability":
			if !a.started {
				return fmt.Errorf("missing load start for %d", id)
			}
			change.HP = -event.Damage
			a.healthChanged = a.healthChanged || change.HP != 0
		case "player_ability":
			if !a.started || cost == nil || heal == nil || *heal != 0 {
				return fmt.Errorf("load fixture requires a recorded damaging starter for %d", id)
			}
			change.MP = -*cost
			a.energyChanged = a.energyChanged || change.MP != 0
			a.npc -= event.Damage
		default:
			return fmt.Errorf("unexpected event in active load encounter: %s", event.Kind)
		}
		if event.Kind != "started" {
			a.replayed, err = character.ChangeResources(a.replayed, change, event.OccurredAt)
			if err != nil {
				return fmt.Errorf("load resource replay %d: %w", id, err)
			}
		}
		hp, mp := a.replayed.Health, a.replayed.Energy
		observedHP, _ := character.RecoverResourceWithRemainder(hp.Current, hp.Maximum, hp.Remainder, hp.UpdatedAt, event.OccurredAt, character.HealthRecoveryPeriod)
		observedMP, _ := character.RecoverResourceWithRemainder(mp.Current, mp.Maximum, mp.Remainder, mp.UpdatedAt, event.OccurredAt, character.EnergyRecoveryPeriod)
		if observedHP != event.CharacterHP || observedMP != event.CharacterMP || a.npc != event.NPCHP {
			return fmt.Errorf("load event/resource mismatch in battle %d sequence %d", id, event.Sequence)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	matches := func(stored, replayed character.Resource, changed bool) bool {
		if stored.Maximum != replayed.Maximum || stored.Current != replayed.Current || stored.Remainder != replayed.Remainder {
			return false
		}
		if changed {
			return stored.UpdatedAt.Equal(replayed.UpdatedAt)
		}
		// Untouched full resources retain their pre-battle initialization time.
		return !stored.UpdatedAt.After(replayed.UpdatedAt)
	}
	for id, a := range audits {
		if !a.started || a.npcStored != a.npc || !matches(a.stored.Health, a.replayed.Health, a.healthChanged) || !matches(a.stored.Energy, a.replayed.Energy, a.energyChanged) {
			return fmt.Errorf("load persisted resources differ from ledger in battle %d", id)
		}
	}
	return nil
}
