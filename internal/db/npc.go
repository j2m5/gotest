package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/character"
	"gotest/internal/npc"
	"gotest/internal/world"
)

type NPCRepository struct{ pool *pgxpool.Pool }

func NewNPCRepository(pool *pgxpool.Pool) *NPCRepository { return &NPCRepository{pool: pool} }

func (r *NPCRepository) List(ctx context.Context, actor character.Actor, expected, locationID int) (*npc.List, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer rollbackCharacterTx(tx)
	if err = authenticateGameActor(ctx, tx, actor, false); err != nil {
		return nil, err
	}
	faction, current, err := worldCharacter(ctx, tx, actor, expected, false)
	if err != nil {
		return nil, err
	}
	if current != locationID {
		return nil, world.ErrPositionChanged
	}
	location, err := scanLocation(tx.QueryRow(ctx, `SELECT `+locationColumns+` FROM locations WHERE id=$1`, current))
	if err != nil {
		return nil, err
	}
	planet, err := worldPlanet(ctx, tx, location.PlanetID)
	if err != nil {
		return nil, err
	}
	if !world.Access(faction, *planet, true).CanTravel {
		return nil, world.ErrAccessDenied
	}
	result := &npc.List{LocationID: locationID, NPCs: []npc.Instance{}}
	rows, err := tx.Query(ctx, `SELECT s.id,t.id,t.name,t.description,t.kind,n.level,n.rank
 FROM npc_spawns s JOIN location_npcs n ON n.id=s.location_npc_id JOIN npc_templates t ON t.id=n.npc_template_id
 WHERE n.location_id=$1 ORDER BY n.position,s.slot,s.id`, locationID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var instance npc.Instance
		if err = rows.Scan(&instance.ID, &instance.NPCTemplateID, &instance.Name, &instance.Description, &instance.Kind, &instance.Level, &instance.Rank); err != nil {
			rows.Close()
			return nil, err
		}
		result.NPCs = append(result.NPCs, instance)
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
