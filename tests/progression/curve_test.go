package progression_test

import (
	"errors"
	"math"
	"testing"

	"gotest/internal/progression"
)

func TestProgressionCumulativeThresholds(t *testing.T) {
	curve := mustCurve(t, []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}, {Number: 3, ExperienceRequired: 250}, {Number: 4, ExperienceRequired: 450}})
	for _, tt := range []struct {
		name   string
		before progression.State
		amount int64
		after  progression.State
	}{
		{"below first threshold", progression.State{Level: 1, Experience: 0}, 99, progression.State{Level: 1, Experience: 99}},
		{"exact second level", progression.State{Level: 1, Experience: 99}, 1, progression.State{Level: 2, Experience: 100}},
		{"below third level", progression.State{Level: 2, Experience: 100}, 149, progression.State{Level: 2, Experience: 249}},
		{"second to third costs 150", progression.State{Level: 2, Experience: 100}, 150, progression.State{Level: 3, Experience: 250}},
		{"several levels with remainder", progression.State{Level: 1, Experience: 0}, 350, progression.State{Level: 3, Experience: 350}},
		{"reach maximum exactly", progression.State{Level: 3, Experience: 250}, 200, progression.State{Level: 4, Experience: 450}},
		{"discard excess at maximum", progression.State{Level: 1, Experience: 50}, math.MaxInt64, progression.State{Level: 4, Experience: 450}},
		{"maximum stops earning", progression.State{Level: 4, Experience: 450}, 1000, progression.State{Level: 4, Experience: 450}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			change, err := curve.Advance(tt.before, tt.amount)
			want := progression.Change{Before: tt.before, After: tt.after, Granted: tt.after.Experience - tt.before.Experience}
			if err != nil || change != want {
				t.Fatalf("Advance = %+v, %v; want %+v", change, err, want)
			}
		})
	}
}

func TestProgressionBalanceChangesPreserveEarnedProgress(t *testing.T) {
	for _, tt := range []struct {
		name   string
		levels []progression.Level
		before progression.State
		amount int64
		after  progression.State
	}{
		{"raised thresholds retain level", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 200}, {Number: 3, ExperienceRequired: 600}}, progression.State{Level: 2, Experience: 100}, 50, progression.State{Level: 2, Experience: 150}},
		{"lowered thresholds allow several levels", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 10}, {Number: 3, ExperienceRequired: 20}, {Number: 4, ExperienceRequired: 50}}, progression.State{Level: 1, Experience: 5}, 20, progression.State{Level: 3, Experience: 25}},
		{"lowered cap retains earned XP", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 10}, {Number: 3, ExperienceRequired: 20}}, progression.State{Level: 2, Experience: 100}, 50, progression.State{Level: 3, Experience: 100}},
		{"raised cap threshold cannot restart max XP", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}, {Number: 3, ExperienceRequired: 500}}, progression.State{Level: 3, Experience: 250}, 50, progression.State{Level: 3, Experience: 250}},
		{"lowered cap retains max XP", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 10}, {Number: 3, ExperienceRequired: 20}}, progression.State{Level: 3, Experience: 250}, 50, progression.State{Level: 3, Experience: 250}},
		{"new level enables new XP", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}, {Number: 3, ExperienceRequired: 250}, {Number: 4, ExperienceRequired: 450}}, progression.State{Level: 3, Experience: 250}, 150, progression.State{Level: 3, Experience: 400}},
		{"one level has no XP", []progression.Level{{Number: 1, ExperienceRequired: 0}}, progression.State{Level: 1, Experience: 0}, 50, progression.State{Level: 1, Experience: 0}},
		{"large thresholds cannot overflow addition", []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: math.MaxInt64}}, progression.State{Level: 1, Experience: math.MaxInt64 - 10}, math.MaxInt64, progression.State{Level: 2, Experience: math.MaxInt64}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			curve := mustCurve(t, tt.levels)
			change, err := curve.Advance(tt.before, tt.amount)
			want := progression.Change{Before: tt.before, After: tt.after, Granted: tt.after.Experience - tt.before.Experience}
			if err != nil || change != want {
				t.Fatalf("Advance = %+v, %v; want %+v", change, err, want)
			}
		})
	}
}

func TestProgressionRejectsInvalidCurveAndState(t *testing.T) {
	for _, levels := range [][]progression.Level{
		nil, {{Number: 2, ExperienceRequired: 0}}, {{Number: 1, ExperienceRequired: 1}}, {{Number: 1, ExperienceRequired: 0}, {Number: 3, ExperienceRequired: 250}}, {{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}, {Number: 2, ExperienceRequired: 250}},
		{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}, {Number: 3, ExperienceRequired: 100}}, {{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: -1}}, {{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 250}, {Number: 3, ExperienceRequired: 100}},
	} {
		if _, err := progression.NewCurve(levels); !errors.Is(err, progression.ErrInvalidCurve) {
			t.Errorf("NewCurve(%v) error = %v", levels, err)
		}
	}
	curve := mustCurve(t, []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}})
	for _, state := range []progression.State{{Level: 0, Experience: 0}, {Level: -1, Experience: 0}, {Level: 3, Experience: 0}, {Level: 1, Experience: -1}} {
		if _, err := curve.Advance(state, 1); !errors.Is(err, progression.ErrInvalidState) {
			t.Errorf("Advance(%v) error = %v", state, err)
		}
	}
	for _, amount := range []int64{0, -1, math.MinInt64} {
		if _, err := curve.Advance(progression.State{Level: 1, Experience: 0}, amount); !errors.Is(err, progression.ErrInvalidReward) {
			t.Errorf("Advance amount %d error = %v", amount, err)
		}
	}
	if _, err := (progression.Curve{}).Advance(progression.State{Level: 1, Experience: 0}, 1); !errors.Is(err, progression.ErrInvalidCurve) {
		t.Errorf("zero curve error = %v", err)
	}
}

func TestProgressionCurveOwnsItsConfiguration(t *testing.T) {
	levels := []progression.Level{{Number: 1, ExperienceRequired: 0}, {Number: 2, ExperienceRequired: 100}, {Number: 3, ExperienceRequired: 250}}
	curve := mustCurve(t, levels)
	levels[1].ExperienceRequired = 200
	levels[2].Number = 9
	change, err := curve.Advance(progression.State{Level: 1, Experience: 0}, 150)
	if err != nil || change.After != (progression.State{Level: 2, Experience: 150}) {
		t.Fatalf("curve changed with input: %+v, %v", change, err)
	}
}

func mustCurve(t *testing.T, levels []progression.Level) progression.Curve {
	t.Helper()
	curve, err := progression.NewCurve(levels)
	if err != nil {
		t.Fatal(err)
	}
	return curve
}
