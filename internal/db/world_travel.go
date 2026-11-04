package db

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/world"
)

const locationColumns = `id,planet_id,code,name,description,position,recommended_level`

func scanLocation(row pgx.Row) (*world.Location, error) {
	l := &world.Location{}
	err := row.Scan(&l.ID, &l.PlanetID, &l.Code, &l.Name, &l.Description, &l.Position, &l.RecommendedLevel)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, world.ErrLocationsUnavailable
	}
	return l, err
}
func worldCharacter(ctx context.Context, tx pgx.Tx, a character.Actor, expected int, lock bool) (int, int, error) {
	if err := requireSelectedCharacter(ctx, tx, a, expected); err != nil {
		return 0, 0, err
	}
	query := `SELECT faction_id,location_id FROM characters WHERE id=$1 AND user_id=$2`
	if lock {
		query += ` FOR UPDATE`
	}
	var faction, location int
	err := tx.QueryRow(ctx, query, expected, a.UserID).Scan(&faction, &location)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, world.ErrCharacterChanged
	}
	return faction, location, err
}
func worldPlanet(ctx context.Context, tx pgx.Tx, id int) (*world.Planet, error) {
	p, err := scanWorldPlanet(tx.QueryRow(ctx, worldPlanetQuery+`WHERE p.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, world.ErrNotFound
	}
	return p, err
}
func worldHome(ctx context.Context, tx pgx.Tx, factionID int) (int, error) {
	var id int
	err := tx.QueryRow(ctx, `SELECT p.id FROM factions f JOIN planets p ON p.code=CASE f.code
	 WHEN 'galactic_republic' THEN 'coruscant' WHEN 'sith_empire' THEN 'dromund_kaas' END
	 WHERE f.id=$1 AND p.core_faction_id=f.id`, factionID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, world.ErrLocationsUnavailable
	}
	return id, err
}

func (r *WorldRepository) Navigation(ctx context.Context, a character.Actor, expected int) (*world.Navigation, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, a, false); err != nil {
		return nil, err
	}
	faction, locationID, err := worldCharacter(ctx, tx, a, expected, false)
	if err != nil {
		return nil, err
	}
	l, err := scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE id=$1`, locationID))
	if err != nil {
		return nil, err
	}
	p, err := worldPlanet(ctx, tx, l.PlanetID)
	if err != nil {
		return nil, err
	}
	catalog, err := worldCatalog(ctx, tx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT DISTINCT planet_id FROM locations`)
	if err != nil {
		return nil, err
	}
	available := map[int]bool{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		available[id] = true
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	result := &world.Navigation{Position: world.Position{CharacterID: expected, Location: *l, Planet: *p}, Destinations: []world.Destination{}, ReturnRequired: !world.Access(faction, *p, true).CanTravel}
	for _, planet := range catalog.Planets {
		result.Destinations = append(result.Destinations, world.Access(faction, planet, available[planet.ID]))
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *WorldRepository) Travel(ctx context.Context, a character.Actor, in world.TravelInput) (*world.TravelResult, error) {
	return r.move(ctx, a, in, 0)
}

// Local and interplanet movement share authentication, locks and reconciliation.
// A nonzero targetLocationID selects a particular area on the current planet.
func (r *WorldRepository) move(ctx context.Context, a character.Actor, in world.TravelInput, targetLocationID int) (*world.TravelResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, a, true); err != nil {
		return nil, err
	}
	faction, locationID, err := worldCharacter(ctx, tx, a, in.CharacterID, true)
	if err != nil {
		return nil, err
	}
	// Hold the current location while determining its planet, preventing content
	// edits from reparenting it beneath the access check.
	current, err := scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE id=$1 FOR SHARE`, locationID))
	if err != nil {
		return nil, err
	}
	var localTarget *world.Location
	if targetLocationID != 0 {
		localTarget, err = scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE id=$1 FOR SHARE`, targetLocationID))
		if errors.Is(err, world.ErrLocationsUnavailable) {
			return nil, world.ErrLocationNotFound
		}
		if err != nil {
			return nil, err
		}
		if localTarget.PlanetID != current.PlanetID {
			return nil, world.ErrWrongPlanet
		}
		in.PlanetID = current.PlanetID
	}
	homeID, err := worldHome(ctx, tx, faction)
	if err != nil {
		return nil, err
	}
	// FOR SHARE also blocks non-key owner updates; FOR KEY SHARE would not.
	rows, err := tx.Query(ctx, `SELECT id FROM planets WHERE id=ANY($1::integer[]) ORDER BY id FOR SHARE`, []int{current.PlanetID, in.PlanetID, homeID})
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// Separate statement after lock acquisition sees any committed control change.
	source, err := worldPlanet(ctx, tx, current.PlanetID)
	if err != nil {
		return nil, err
	}
	returnHome := !world.Access(faction, *source, true).CanTravel
	targetID := in.PlanetID
	if returnHome {
		targetID = homeID
	}
	target, err := worldPlanet(ctx, tx, targetID)
	if err != nil {
		return nil, err
	}
	if !world.Access(faction, *target, true).CanTravel {
		return nil, world.ErrAccessDenied
	}
	arrival := current
	if localTarget != nil && !returnHome {
		arrival = localTarget
	} else if current.PlanetID != targetID {
		arrival, err = scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE planet_id=$1 ORDER BY recommended_level,position,id LIMIT 1 FOR SHARE`, targetID))
		if err != nil {
			return nil, err
		}
	}
	if arrival.ID != current.ID && current.ID != in.FromLocationID {
		return nil, world.ErrPositionChanged
	}
	// Locks may outlive the session even when no other transaction can log out.
	if err = authenticateGameActor(ctx, tx, a, false); err != nil {
		return nil, err
	}
	if arrival.ID != current.ID {
		if _, err = tx.Exec(ctx, `UPDATE characters SET location_id=$2 WHERE id=$1`, in.CharacterID, arrival.ID); err != nil {
			return nil, err
		}
	}
	result := &world.TravelResult{Position: world.Position{CharacterID: in.CharacterID, Location: *arrival, Planet: *target}, ReturnedHome: returnHome}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
