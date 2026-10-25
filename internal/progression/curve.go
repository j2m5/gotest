package progression

import "slices"

// Curve is an immutable, contiguous catalog of cumulative experience thresholds.
type Curve struct{ levels []Level }

func NewCurve(levels []Level) (Curve, error) {
	if len(levels) == 0 || levels[0].Number != 1 || levels[0].ExperienceRequired != 0 {
		return Curve{}, ErrInvalidCurve
	}
	for i, level := range levels {
		if int(level.Number) != i+1 || (i > 0 && level.ExperienceRequired <= levels[i-1].ExperienceRequired) {
			return Curve{}, ErrInvalidCurve
		}
	}
	return Curve{levels: slices.Clone(levels)}, nil
}

// Advance discards excess XP at the cap. Rebalancing may leave an earned level
// or experience above its new threshold; a new reward never removes either.
func (c Curve) Advance(current State, amount int64) (Change, error) {
	if len(c.levels) == 0 {
		return Change{}, ErrInvalidCurve
	}
	if current.Level <= 0 || int(current.Level) > len(c.levels) || current.Experience < 0 {
		return Change{}, ErrInvalidState
	}
	if amount <= 0 {
		return Change{}, ErrInvalidReward
	}
	after := current
	maximum := c.levels[len(c.levels)-1]
	if current.Level < maximum.Number {
		// Clamp before adding so even a MaxInt64 reward cannot overflow.
		remaining := max(int64(0), maximum.ExperienceRequired-current.Experience)
		after.Experience += min(amount, remaining)
		for _, level := range c.levels {
			if level.ExperienceRequired <= after.Experience && level.Number > after.Level {
				after.Level = level.Number
			}
		}
	}
	return Change{Before: current, After: after, Granted: after.Experience - current.Experience}, nil
}
