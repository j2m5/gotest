package character_test

import (
	"math"
	"testing"
	"time"

	"gotest/internal/character"
)

func TestResourceRecovery(t *testing.T) {
	at := time.Date(2026, 10, 29, 12, 0, 0, 123456000, time.UTC)
	for _, tc := range []struct {
		name              string
		baseline, maximum int
		elapsed, period   time.Duration
		want              int
	}{
		{"health starts empty", 0, 100, 0, 5 * time.Minute, 0},
		{"health half", 0, 100, 150 * time.Second, 5 * time.Minute, 50},
		{"health just before full", 0, 100, 5*time.Minute - time.Millisecond, 5 * time.Minute, 99},
		{"health full", 0, 100, 5 * time.Minute, 5 * time.Minute, 100},
		{"partial takes less time", 75, 100, 75 * time.Second, 5 * time.Minute, 100},
		{"energy half", 0, 100, 30 * time.Second, time.Minute, 50},
		{"energy just before full", 0, 100, time.Minute - time.Millisecond, time.Minute, 99},
		{"energy full", 0, 100, time.Minute, time.Minute, 100},
		{"zero energy capacity", 0, 0, time.Hour, time.Minute, 0},
		{"odd capacity", 0, 7, 150 * time.Second, 5 * time.Minute, 3},
		{"fraction not rounded up", 37, 200, time.Second, 5 * time.Minute, 37},
		{"future anchor", 20, 100, -time.Second, 5 * time.Minute, 20},
		{"offline for a year", 0, 100, 365 * 24 * time.Hour, 5 * time.Minute, 100},
		{"int32 maximum", 0, math.MaxInt32, 150 * time.Second, 5 * time.Minute, math.MaxInt32 / 2},
		{"near maximum no sum overflow", math.MaxInt32 - 1, math.MaxInt32, time.Second, 5 * time.Minute, math.MaxInt32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := character.RecoverResource(tc.baseline, tc.maximum, at, at.Add(tc.elapsed), tc.period); got != tc.want {
				t.Fatalf("recovery=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestResourceRecoveryKeepsFractionBetweenReads(t *testing.T) {
	at := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	for second := 0; second <= 6; second++ {
		got := character.RecoverResource(0, 100, at, at.Add(time.Duration(second)*time.Second), character.HealthRecoveryPeriod)
		if got != second/3 {
			t.Fatalf("second %d: got %d", second, got)
		}
	}
	if character.HealthRecoveryPeriod != 5*time.Minute || character.EnergyRecoveryPeriod != time.Minute {
		t.Fatal("unexpected recovery balance")
	}
}
