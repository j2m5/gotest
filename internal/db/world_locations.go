package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gotest/internal/character"
	"gotest/internal/world"
)

func (r *WorldRepository) Locations(ctx context.Context, actor character.Actor, expected, planetID int) (*world.LocationCatalog, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	faction, locationID, err := worldCharacter(ctx, tx, actor, expected, false)
	if err != nil {
		return nil, err
	}
	current, err := scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE id=$1`, locationID))
	if err != nil {
		return nil, err
	}
	if current.PlanetID != planetID {
		return nil, world.ErrWrongPlanet
	}
	planet, err := worldPlanet(ctx, tx, planetID)
	if err != nil {
		return nil, err
	}
	result := &world.LocationCatalog{Position: world.Position{CharacterID: expected, Location: *current, Planet: *planet}, Locations: []world.Location{}, ReturnRequired: !world.Access(faction, *planet, true).CanTravel}
	rows, err := tx.Query(ctx, `SELECT `+locationColumns+` FROM locations WHERE planet_id=$1 ORDER BY position,id`, planetID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		location, err := scanLocation(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		result.Locations = append(result.Locations, *location)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *WorldRepository) Move(ctx context.Context, actor character.Actor, in world.MoveInput) (*world.TravelResult, error) {
	return r.move(ctx, actor, world.TravelInput{CharacterID: in.CharacterID, FromLocationID: in.FromLocationID}, in.LocationID)
}
