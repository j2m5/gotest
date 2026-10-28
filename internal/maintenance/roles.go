package maintenance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
)

type RoleChange struct {
	Previous, Role string
	Changed        bool
}

func ValidAccountRole(role string) bool {
	return role == "user" || role == "moderator" || role == "administrator"
}

// SetAccountRole is a privileged local maintenance operation, never an HTTP
// endpoint. It shares the user lock with moderation and records the CLI actor
// as NULL so the audit cannot impersonate an in-game administrator.
func SetAccountRole(ctx context.Context, pool *pgxpool.Pool, userID int, role string) (RoleChange, error) {
	var result RoleChange
	if userID <= 0 || !ValidAccountRole(role) {
		return result, errors.New("invalid user ID or account role")
	}
	roleID := auth.RoleUser
	if role == "moderator" {
		roleID = auth.RoleModerator
	}
	if role == "administrator" {
		roleID = auth.RoleAdministrator
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	var previous int16
	if err = tx.QueryRow(ctx, `SELECT role_id FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&previous); err != nil {
		return result, err
	}
	result = RoleChange{Previous: auth.RoleName(previous), Role: role, Changed: previous != roleID}
	if result.Changed {
		if _, err = tx.Exec(ctx, `UPDATE users SET role_id=$2,updated_at=clock_timestamp() WHERE id=$1`, userID, roleID); err != nil {
			return RoleChange{}, err
		}
		reason := fmt.Sprintf("Local role change: %s -> %s", result.Previous, role)
		if _, err = tx.Exec(ctx, `INSERT INTO moderation_actions(user_id,scope,action,reason) VALUES($1,'account','role_change',$2)`, userID, reason); err != nil {
			return RoleChange{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return RoleChange{}, err
	}
	return result, nil
}
