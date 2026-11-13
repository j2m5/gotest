package character

import (
	"math/bits"
	"time"
)

const (
	HealthRecoveryPeriod = 5 * time.Minute
	EnergyRecoveryPeriod = time.Minute
)

// Regeneration describes the persisted baselines behind a server-calculated
// character view. ObservedAt uses the database clock, never a client timestamp.
type Regeneration struct {
	HPBaseline      int       `json:"hp_baseline"`
	MPBaseline      int       `json:"mp_baseline"`
	HPRemainder     int64     `json:"hp_remainder"`
	MPRemainder     int64     `json:"mp_remainder"`
	HPUpdatedAt     time.Time `json:"hp_updated_at"`
	MPUpdatedAt     time.Time `json:"mp_updated_at"`
	ObservedAt      time.Time `json:"observed_at"`
	HPFullRestoreMS int64     `json:"hp_full_restore_ms"`
	MPFullRestoreMS int64     `json:"mp_full_restore_ms"`
}

// RecoverResource projects an int32 database resource without modifying its
// baseline. Millisecond precision matches browser Date; retaining the original
// anchor means frequent reads cannot discard fractional recovery.
func RecoverResource(baseline, maximum int, updatedAt, observedAt time.Time, period time.Duration) int {
	current, _ := RecoverResourceWithRemainder(baseline, maximum, 0, updatedAt, observedAt, period)
	return current
}

// RecoverResourceWithRemainder retains the numerator of a fractional resource
// point. The denominator is period.Milliseconds(); full resources keep no carry.
func RecoverResourceWithRemainder(baseline, maximum int, remainder int64, updatedAt, observedAt time.Time, period time.Duration) (int, int64) {
	if maximum <= 0 {
		return 0, 0
	}
	if baseline < 0 {
		baseline = 0
	}
	if baseline >= maximum {
		return maximum, 0
	}
	elapsed := observedAt.Truncate(time.Millisecond).Sub(updatedAt.Truncate(time.Millisecond)).Milliseconds()
	duration := period.Milliseconds()
	if duration <= 0 {
		return baseline, 0
	}
	if remainder < 0 || remainder >= duration {
		remainder = 0
	}
	if elapsed <= 0 {
		return baseline, remainder
	}
	if elapsed >= duration {
		return maximum, 0
	}
	// Keep multiplication exact even for capacities/periods beyond game defaults.
	hi, lo := bits.Mul64(uint64(maximum), uint64(elapsed))
	lo, carry := bits.Add64(lo, uint64(remainder), 0)
	recovered, rest := bits.Div64(hi+carry, lo, uint64(duration))
	if recovered >= uint64(maximum-baseline) {
		return maximum, 0
	}
	return baseline + int(recovered), int64(rest)
}
