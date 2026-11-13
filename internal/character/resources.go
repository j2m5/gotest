package character

import (
	"errors"
	"math"
	"time"
)

var (
	ErrInsufficientEnergy    = errors.New("character: insufficient energy")
	ErrInvalidResourceChange = errors.New("character: invalid resource change")
)

type Resource struct {
	Current, Maximum int
	Remainder        int64
	UpdatedAt        time.Time
}

type Resources struct{ Health, Energy Resource }
type ResourceChange struct{ HP, MP int }

// ChangeResources settles only changed resources. Whole-point effects preserve
// fractional recovery; saturation clears it. Energy spending must be affordable.
func ChangeResources(resources Resources, change ResourceChange, at time.Time) (Resources, error) {
	valid := func(r Resource, period time.Duration) bool {
		return r.Maximum >= 0 && r.Maximum <= math.MaxInt32 && r.Current >= 0 && r.Current <= r.Maximum &&
			r.Remainder >= 0 && r.Remainder < period.Milliseconds() && (r.Current < r.Maximum || r.Remainder == 0)
	}
	if !valid(resources.Health, HealthRecoveryPeriod) || !valid(resources.Energy, EnergyRecoveryPeriod) ||
		int64(change.HP) < -math.MaxInt32 || int64(change.HP) > math.MaxInt32 || int64(change.MP) < -math.MaxInt32 || int64(change.MP) > math.MaxInt32 {
		return resources, ErrInvalidResourceChange
	}
	apply := func(r Resource, delta int, period time.Duration, requireAffordable bool) (Resource, error) {
		if delta == 0 {
			return r, nil
		}
		r.Current, r.Remainder = RecoverResourceWithRemainder(r.Current, r.Maximum, r.Remainder, r.UpdatedAt, at, period)
		value := int64(r.Current) + int64(delta)
		if value < 0 && requireAffordable {
			return Resource{}, ErrInsufficientEnergy
		}
		if value < 0 {
			value, r.Remainder = 0, 0
		}
		if value >= int64(r.Maximum) {
			value, r.Remainder = int64(r.Maximum), 0
		}
		r.Current = int(value)
		if at.After(r.UpdatedAt) {
			r.UpdatedAt = at.Truncate(time.Millisecond)
		}
		return r, nil
	}
	next := resources
	var err error
	if next.Energy, err = apply(resources.Energy, change.MP, EnergyRecoveryPeriod, true); err != nil {
		return resources, err
	}
	next.Health, err = apply(resources.Health, change.HP, HealthRecoveryPeriod, false)
	if err != nil {
		return resources, err
	}
	return next, nil
}
