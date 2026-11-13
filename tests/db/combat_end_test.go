package db_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/combat"
	"gotest/internal/db"
	"gotest/internal/wallet"
	"gotest/internal/world"
)

func TestCombatRetreatRestoresNPCAndPreservesPlayer(t *testing.T) {
	s, ctx, actor, observer, in := combatFixture(t)
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=50,mp_current=20 WHERE id=2;
	 UPDATE characters SET global_cooldown_ready_at=clock_timestamp()+interval '1 hour' WHERE id=2`)
	r := db.NewCombatRepository(s.pool)
	b := startBattle(t, ctx, s, actor, in).Battle
	mustExec(t, ctx, s, `UPDATE battles SET npc_hp_current=7 WHERE id=$1`, b.ID)
	before := characterStatsSnapshot(t, ctx, s)
	for _, viewer := range []struct {
		actor       character.Actor
		characterID int
	}{{actor, 2}, {observer, 3}} {
		list, err := db.NewNPCRepository(s.pool).List(ctx, viewer.actor, viewer.characterID, in.LocationID)
		if err != nil || len(list.NPCs) != 3 {
			t.Fatalf("occupied NPC visible: %+v %v", list, err)
		}
	}
	moves := db.NewWorldRepository(s.pool)
	_, err := moves.Move(ctx, actor, world.MoveInput{CharacterID: 2, LocationID: locationAt(t, ctx, s, "coruscant", 2), FromLocationID: in.LocationID})
	if !errors.Is(err, combat.ErrInBattle) {
		t.Fatalf("move during battle: %v", err)
	}
	_, err = moves.Travel(ctx, actor, world.TravelInput{CharacterID: 2, PlanetID: planetID(t, ctx, s, "tython"), FromLocationID: in.LocationID})
	if !errors.Is(err, combat.ErrInBattle) {
		t.Fatalf("travel during battle: %v", err)
	}
	retreated, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.RetreatInTx(ctx, tx, actor, 2, b.ID) })
	if err != nil || retreated.Replayed || retreated.Battle.Status != combat.Retreated || retreated.Battle.NPC.HPCurrent != 200 || retreated.Battle.EndedAt == nil {
		t.Fatalf("retreat: %+v %v", retreated, err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("retreat changed player resources or cooldowns")
	}
	lateDefeat, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.DefeatInTx(ctx, tx, 2, b.ID) })
	if err != nil || !lateDefeat.Replayed || lateDefeat.Battle.Status != combat.Retreated || characterStatsSnapshot(t, ctx, s) != before {
		t.Fatalf("late defeat after retreat: %+v %v", lateDefeat, err)
	}
	state, err := r.Current(ctx, actor, 2)
	if err != nil || state.Battle != nil {
		t.Fatalf("active after retreat: %+v %v", state, err)
	}
	list, err := db.NewNPCRepository(s.pool).List(ctx, observer, 3, in.LocationID)
	if err != nil || len(list.NPCs) != 4 {
		t.Fatalf("NPC not released: %+v %v", list, err)
	}
	// A lost response to start must not resurrect the now finished encounter.
	if replay := startBattle(t, ctx, s, actor, in); !replay.Replayed || replay.Battle.Status != combat.Retreated {
		t.Fatal("start replay reopened battle")
	}
	if _, err = moves.Move(ctx, actor, world.MoveInput{CharacterID: 2, LocationID: locationAt(t, ctx, s, "coruscant", 2), FromLocationID: in.LocationID}); err != nil {
		t.Fatal(err)
	}
}

func TestCombatOldRetreatCannotReleaseANewBattle(t *testing.T) {
	s, ctx, actor, other, in := combatFixture(t)
	r := db.NewCombatRepository(s.pool)
	old := startBattle(t, ctx, s, actor, in).Battle
	if _, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.RetreatInTx(ctx, tx, actor, 2, old.ID) }); err != nil {
		t.Fatal(err)
	}
	in.CharacterID = 3
	in.Key = "next-player"
	newBattle := startBattle(t, ctx, s, other, in).Battle
	mustExec(t, ctx, s, `UPDATE battles SET npc_hp_current=90 WHERE id=$1`, newBattle.ID)
	replay, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.RetreatInTx(ctx, tx, actor, 2, old.ID) })
	if err != nil || !replay.Replayed {
		t.Fatalf("retreat replay: %+v %v", replay, err)
	}
	current, err := r.Current(ctx, other, 3)
	if err != nil || current.Battle == nil || current.Battle.ID != newBattle.ID || current.Battle.NPC.HPCurrent != 90 {
		t.Fatalf("new battle affected: %+v %v", current, err)
	}
	_, err = combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.RetreatInTx(ctx, tx, actor, 2, newBattle.ID) })
	if !errors.Is(err, combat.ErrNotFound) {
		t.Fatalf("foreign battle ended: %v", err)
	}
}

func TestCombatDefeatAndRollback(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=40,mp_current=31,experience=37 WHERE id=2`)
	for _, currency := range []string{wallet.Credits, wallet.CartelCoins} {
		if _, err := db.NewWalletRepository(s.pool).Apply(ctx, walletOperation(2, currency, 123, "before-defeat")); err != nil {
			t.Fatal(err)
		}
	}
	b := startBattle(t, ctx, s, actor, in).Battle
	mustExec(t, ctx, s, `UPDATE battles SET npc_hp_current=3 WHERE id=$1`, b.ID)
	r := db.NewCombatRepository(s.pool)
	before := characterStatsSnapshot(t, ctx, s)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	result, err := r.DefeatInTx(ctx, tx, 2, b.ID)
	if err != nil || result.Battle.Status != combat.Defeated || result.Battle.NPC.HPCurrent != 200 {
		t.Fatalf("defeat: %+v %v", result, err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if characterStatsSnapshot(t, ctx, s) != before {
		t.Fatal("rollback lost player state")
	}
	current, err := r.Current(ctx, actor, 2)
	if err != nil || current.Battle == nil || current.Battle.NPC.HPCurrent != 3 {
		t.Fatalf("rollback lost battle: %+v %v", current, err)
	}
	_, err = combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.DefeatInTx(ctx, tx, 2, b.ID) })
	if err != nil {
		t.Fatal(err)
	}
	var hp, mp, location, xp int
	if err = s.pool.QueryRow(ctx, `SELECT hp_current,mp_current,location_id,experience FROM characters WHERE id=2`).Scan(&hp, &mp, &location, &xp); err != nil || hp != 0 || mp != 31 || location != in.LocationID || xp != 37 {
		t.Fatalf("defeat player: hp=%d mp=%d loc=%d xp=%d %v", hp, mp, location, xp, err)
	}
	mustExec(t, ctx, s, `UPDATE characters SET hp_current=30 WHERE id=2`)
	replay, err := combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.DefeatInTx(ctx, tx, 2, b.ID) })
	if err != nil || !replay.Replayed {
		t.Fatalf("defeat replay: %+v %v", replay, err)
	}
	if err = s.pool.QueryRow(ctx, `SELECT hp_current FROM characters WHERE id=2`).Scan(&hp); err != nil || hp != 30 {
		t.Fatal("defeat replay reset recovered health")
	}
	for _, currency := range []string{wallet.Credits, wallet.CartelCoins} {
		if balance, count := walletState(t, ctx, s, 2, currency); balance != 123 || count != 1 {
			t.Fatalf("defeat changed %s: %d / %d", currency, balance, count)
		}
	}
}

func TestCombatSurvivesLogoutSelectionAndRepositoryRestart(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	b := startBattle(t, ctx, s, actor, in).Battle
	chars := db.NewCharacterRepository(s.pool)
	if _, err := chars.SetActive(ctx, actor, new(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.NewCombatRepository(s.pool).Current(ctx, actor, 2); !errors.Is(err, character.ErrSelectionChanged) {
		t.Fatal(err)
	}
	if err := s.DeleteSession(ctx, actor.SessionHash); err != nil {
		t.Fatal(err)
	}
	actor = repositoryActor(t, ctx, s, 1, "combat-after-login")
	if _, err := chars.SetActive(ctx, actor, new(2)); err != nil {
		t.Fatal(err)
	}
	state, err := db.NewCombatRepository(s.pool).Current(ctx, actor, 2)
	if err != nil || state.Battle == nil || state.Battle.ID != b.ID {
		t.Fatalf("lost battle after login: %+v %v", state, err)
	}
}

func TestCombatLostPlanetAllowsRetreatButBlocksForcedTravel(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	in.LocationID = locationAt(t, ctx, s, "balmorra", 1)
	mustExec(t, ctx, s, `UPDATE characters SET location_id=$1 WHERE id=2`, in.LocationID)
	if err := s.pool.QueryRow(ctx, `SELECT s.id FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id WHERE n.location_id=$1 LIMIT 1`, in.LocationID).Scan(&in.NPCSpawnID); err != nil {
		t.Fatal(err)
	}
	b := startBattle(t, ctx, s, actor, in).Battle
	mustExec(t, ctx, s, `UPDATE planets SET controlling_faction_id=NULL WHERE code='balmorra'`)
	r := db.NewCombatRepository(s.pool)
	state, err := r.Current(ctx, actor, 2)
	if err != nil || state.Battle == nil {
		t.Fatalf("lost planet hid owned battle: %+v %v", state, err)
	}
	travel := world.TravelInput{CharacterID: 2, PlanetID: planetID(t, ctx, s, "coruscant"), FromLocationID: in.LocationID}
	if _, err = db.NewWorldRepository(s.pool).Travel(ctx, actor, travel); !errors.Is(err, combat.ErrInBattle) {
		t.Fatalf("forced move during battle: %v", err)
	}
	if _, err = combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) { return r.RetreatInTx(ctx, tx, actor, 2, b.ID) }); err != nil {
		t.Fatal(err)
	}
	in.Key = "lost-access"
	if _, err = combatTransaction(ctx, s, func(tx pgx.Tx) (*combat.Change, error) {
		return r.StartInTx(ctx, tx, actor, in, combat.NPCStats{HP: 200})
	}); !errors.Is(err, world.ErrAccessDenied) {
		t.Fatalf("start after lost control: %v", err)
	}
	returned, err := db.NewWorldRepository(s.pool).Travel(ctx, actor, travel)
	if err != nil || !returned.ReturnedHome || returned.Position.Planet.Code != "coruscant" {
		t.Fatalf("return after retreat: %+v %v", returned, err)
	}
}

func TestCombatRespawningNPCIsHiddenUntilDeadline(t *testing.T) {
	s, ctx, actor, _, in := combatFixture(t)
	mustExec(t, ctx, s, `UPDATE npc_spawns SET respawn_at=clock_timestamp()+interval '1 hour' WHERE id=$1`, in.NPCSpawnID)
	list, err := db.NewNPCRepository(s.pool).List(ctx, actor, 2, in.LocationID)
	if err != nil || len(list.NPCs) != 3 {
		t.Fatalf("respawning NPC visible: %+v %v", list, err)
	}
	mustExec(t, ctx, s, `UPDATE npc_spawns SET respawn_at=clock_timestamp()-interval '1 second' WHERE id=$1`, in.NPCSpawnID)
	list, err = db.NewNPCRepository(s.pool).List(ctx, actor, 2, in.LocationID)
	if err != nil || len(list.NPCs) != 4 {
		t.Fatalf("expired respawn still hidden: %+v %v", list, err)
	}
	startBattle(t, ctx, s, actor, in)
}
