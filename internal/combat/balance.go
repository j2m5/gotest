package combat

import "math"

const MaxBattleMilliseconds int64 = 600000

// Rules is an immutable per-encounter snapshot. Changing the provisional balance
// below affects new encounters only; it is never supplied by a player's request.
type Rules struct {
	Version    int    `json:"version"`
	HP         int    `json:"hp"`
	Defense    int    `json:"defense"`
	Experience int64  `json:"experience"`
	Credits    int64  `json:"credits"`
	RespawnMS  int64  `json:"respawn_ms"`
	DurationMS int64  `json:"duration_ms"`
	Pattern    []Step `json:"pattern"`
}

type Step struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Damage  int    `json:"damage"`
	DelayMS int64  `json:"delay_ms"`
}

func RulesFor(level int, rank string) (Rules, error) {
	multiplier := map[string]int{"normal": 1, "strong": 2, "elite": 3}[rank]
	if level < 1 || level > 50 || multiplier == 0 {
		return Rules{}, ErrInvalidRules
	}
	base := (4 + 2*level) * multiplier
	normal := Step{Code: "npc_strike", Name: "Обычный удар", Damage: base, DelayMS: 2000}
	heavy := Step{Code: "npc_heavy_strike", Name: "Сильный удар", Damage: 2 * base, DelayMS: 3000}
	pattern := []Step{normal, normal, heavy}
	if rank == "strong" {
		pattern = []Step{normal, heavy}
	}
	if rank == "elite" {
		pattern = []Step{heavy, normal, heavy}
	}
	return Rules{Version: 1, HP: (48 + 12*level) * multiplier, Defense: (level - 1) / 2,
		Experience: int64(5 * level * multiplier), Credits: int64(3 * level * multiplier),
		RespawnMS: 30000, DurationMS: MaxBattleMilliseconds, Pattern: pattern}, nil
}

func (r Rules) Validate() error {
	if r.Version != 1 || !ValidID(r.HP) || r.Defense < 0 || r.Defense > math.MaxInt32 ||
		r.Experience <= 0 || r.Experience > math.MaxInt32 || r.Credits <= 0 || r.Credits > math.MaxInt32 ||
		r.RespawnMS < 1000 || r.RespawnMS > 3600000 || r.DurationMS < 1000 || r.DurationMS > MaxBattleMilliseconds ||
		len(r.Pattern) < 1 || len(r.Pattern) > 16 {
		return ErrInvalidRules
	}
	for _, step := range r.Pattern {
		if len(step.Code) == 0 || len(step.Code) > 64 || len(step.Name) == 0 || len(step.Name) > 300 || !ValidID(step.Damage) || step.DelayMS < 1000 || step.DelayMS > 60000 {
			return ErrInvalidRules
		}
	}
	return nil
}

// Damage keeps intermediate arithmetic in int64 and bounds the persisted result.
// A non-damaging ability never acquires implicit damage from the attacker's stats.
func Damage(base, power, mastery, defense int) int {
	if base <= 0 {
		return 0
	}
	amount := int64(base) + int64(max(power, 0))/5 + int64(max(mastery, 0))/10
	amount = amount * 100 / (100 + int64(max(defense, 0)))
	return int(max(1, min(amount, math.MaxInt32)))
}
