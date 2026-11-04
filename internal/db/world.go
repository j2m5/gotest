package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/world"
)

type WorldRepository struct{ pool *pgxpool.Pool }

func NewWorldRepository(pool *pgxpool.Pool) *WorldRepository { return &WorldRepository{pool: pool} }

const worldPlanetQuery = `SELECT p.id,p.code,p.name,p.name_en,p.description,
	 cf.id,COALESCE(cf.code,''),COALESCE(NULLIF(btrim(cf.alias),''),cf.name,''),
	 f.id,COALESCE(f.code,''),COALESCE(NULLIF(btrim(f.alias),''),f.name,'')
	 FROM planets p LEFT JOIN factions cf ON cf.id=p.core_faction_id
	 LEFT JOIN factions f ON f.id=p.controlling_faction_id `

type worldQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func scanWorldPlanet(row pgx.Row) (*world.Planet, error) {
	var planet world.Planet
	var coreID, controllerID *int
	var core, controller world.Faction
	if err := row.Scan(&planet.ID, &planet.Code, &planet.Name, &planet.NameEN, &planet.Description,
		&coreID, &core.Code, &core.Name, &controllerID, &controller.Code, &controller.Name); err != nil {
		return nil, err
	}
	if coreID != nil {
		core.ID = *coreID
		planet.CoreFaction = &core
	}
	if controllerID != nil {
		controller.ID = *controllerID
		planet.ControllingFaction = &controller
	}
	return &planet, nil
}

func (r *WorldRepository) Catalog(ctx context.Context) (*world.Catalog, error) {
	return worldCatalog(ctx, r.pool)
}
func worldCatalog(ctx context.Context, q worldQueryer) (*world.Catalog, error) {
	rows, err := q.Query(ctx, worldPlanetQuery+`ORDER BY p.name,p.code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := &world.Catalog{Planets: []world.Planet{}}
	for rows.Next() {
		planet, err := scanWorldPlanet(rows)
		if err != nil {
			return nil, err
		}
		result.Planets = append(result.Planets, *planet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
