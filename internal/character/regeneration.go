package character

import "time"

const (
	HealthRecoveryPeriod = 5 * time.Minute
	EnergyRecoveryPeriod = time.Minute
)

// Regeneration describes the persisted baselines behind a server-calculated
// character view. ObservedAt uses the database clock, never a client timestamp.
type Regeneration struct {
	HPBaseline      int       `json:"hp_baseline"`
	MPBaseline      int       `json:"mp_baseline"`
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
	if maximum <= 0 {
		return 0
	}
	if baseline < 0 {
		baseline = 0
	}
	if baseline >= maximum {
		return maximum
	}
	elapsed := observedAt.Truncate(time.Millisecond).Sub(updatedAt.Truncate(time.Millisecond)).Milliseconds()
	duration := period.Milliseconds()
	if elapsed <= 0 || duration <= 0 {
		return baseline
	}
	// Bound elapsed before multiplication: int32 maximum * 300000ms fits int64.
	if elapsed >= duration {
		return maximum
	}
	recovered := int(int64(maximum) * elapsed / duration)
	if recovered >= maximum-baseline {
		return maximum
	}
	return baseline + recovered
}
