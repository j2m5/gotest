package combat

import (
	"math"
	"time"

	"gotest/internal/character"
)

type Event struct {
	Sequence    int       `json:"sequence"`
	Kind        string    `json:"kind"`
	Actor       string    `json:"actor"`
	AbilityCode string    `json:"ability_code"`
	AbilityName string    `json:"ability_name"`
	Damage      int       `json:"damage"`
	Healing     int       `json:"healing"`
	CharacterHP int       `json:"character_hp"`
	CharacterMP int       `json:"character_mp"`
	NPCHP       int       `json:"npc_hp"`
	Experience  int64     `json:"experience,string"`
	Credits     int64     `json:"credits,string"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// Frame is owned by one locked transaction. Advance is deterministic and does no
// IO; missed ticks and live ticks use exactly the same resource arithmetic.
type Frame struct {
	Battle       Battle
	Rules        Rules
	Resources    character.Resources
	Defense      int
	PatternStep  int
	NextActionAt time.Time
	ExpiresAt    time.Time
	Sequence     int
}

func (f *Frame) Record(event Event) Event {
	f.Sequence++
	event.Sequence = f.Sequence
	event.CharacterHP, _ = character.RecoverResourceWithRemainder(f.Resources.Health.Current, f.Resources.Health.Maximum, f.Resources.Health.Remainder, f.Resources.Health.UpdatedAt, event.OccurredAt, character.HealthRecoveryPeriod)
	event.CharacterMP, _ = character.RecoverResourceWithRemainder(f.Resources.Energy.Current, f.Resources.Energy.Maximum, f.Resources.Energy.Remainder, f.Resources.Energy.UpdatedAt, event.OccurredAt, character.EnergyRecoveryPeriod)
	event.NPCHP = f.Battle.NPC.HPCurrent
	return event
}

func (f *Frame) Finish(status Status, at time.Time) {
	f.Battle.Status, f.Battle.EndedAt, f.NextActionAt = status, &at, time.Time{}
	if status != Won {
		f.Battle.NPC.HPCurrent = f.Battle.NPC.HP
	}
}

func Advance(frame Frame, now time.Time) (Frame, []Event, error) {
	if frame.Battle.Status != Active {
		return frame, nil, nil
	}
	if err := frame.Rules.Validate(); err != nil {
		return frame, nil, err
	}
	if frame.NextActionAt.IsZero() || !frame.NextActionAt.After(frame.Battle.CreatedAt) ||
		!frame.ExpiresAt.After(frame.Battle.CreatedAt) || frame.ExpiresAt.Sub(frame.Battle.CreatedAt) > time.Duration(MaxBattleMilliseconds)*time.Millisecond ||
		frame.PatternStep < 0 || frame.PatternStep > math.MaxInt32-601 || frame.Sequence < 0 || frame.Sequence > math.MaxInt32-602 || frame.Defense < 0 || frame.Defense > math.MaxInt32 {
		return frame, nil, ErrInvalidState
	}
	f := frame
	var events []Event
	for count := 0; count <= 600; count++ {
		if !now.Before(f.ExpiresAt) && !f.ExpiresAt.After(f.NextActionAt) {
			f.Finish(Expired, f.ExpiresAt)
			events = append(events, f.Record(Event{Kind: "expired", Actor: "system", OccurredAt: f.ExpiresAt}))
			return f, events, nil
		}
		if now.Before(f.NextActionAt) {
			return f, events, nil
		}
		at := f.NextActionAt
		step := f.Rules.Pattern[f.PatternStep%len(f.Rules.Pattern)]
		before, _ := character.RecoverResourceWithRemainder(f.Resources.Health.Current, f.Resources.Health.Maximum, f.Resources.Health.Remainder, f.Resources.Health.UpdatedAt, at, character.HealthRecoveryPeriod)
		var err error
		f.Resources, err = character.ChangeResources(f.Resources, character.ResourceChange{HP: -Damage(step.Damage, 0, 0, f.Defense)}, at)
		if err != nil {
			return frame, nil, err
		}
		events = append(events, f.Record(Event{Kind: "npc_ability", Actor: "npc", AbilityCode: step.Code, AbilityName: step.Name, Damage: before - f.Resources.Health.Current, OccurredAt: at}))
		f.PatternStep++
		if f.Resources.Health.Current == 0 {
			f.Finish(Defeated, at)
			events = append(events, f.Record(Event{Kind: "defeated", Actor: "system", OccurredAt: at}))
			return f, events, nil
		}
		f.NextActionAt = at.Add(time.Duration(f.Rules.Pattern[f.PatternStep%len(f.Rules.Pattern)].DelayMS) * time.Millisecond)
	}
	return frame, nil, ErrInvalidState
}

func PlayerEffect(frame Frame, damage, heal, power, mastery int, code, name string, at time.Time) (Frame, Event, error) {
	if frame.Battle.Status != Active {
		return frame, Event{}, ErrFinished
	}
	f := frame
	actualDamage := min(f.Battle.NPC.HPCurrent, Damage(damage, power, mastery, f.Rules.Defense))
	f.Battle.NPC.HPCurrent -= actualDamage
	before, _ := character.RecoverResourceWithRemainder(f.Resources.Health.Current, f.Resources.Health.Maximum, f.Resources.Health.Remainder, f.Resources.Health.UpdatedAt, at, character.HealthRecoveryPeriod)
	var err error
	f.Resources, err = character.ChangeResources(f.Resources, character.ResourceChange{HP: Damage(heal, power, mastery, 0)}, at)
	if err != nil {
		return frame, Event{}, err
	}
	after, _ := character.RecoverResourceWithRemainder(f.Resources.Health.Current, f.Resources.Health.Maximum, f.Resources.Health.Remainder, f.Resources.Health.UpdatedAt, at, character.HealthRecoveryPeriod)
	if f.Battle.NPC.HPCurrent == 0 {
		f.Finish(Won, at)
	}
	event := f.Record(Event{Kind: "player_ability", Actor: "player", AbilityCode: code, AbilityName: name, Damage: actualDamage, Healing: after - before, OccurredAt: at})
	return f, event, nil
}
