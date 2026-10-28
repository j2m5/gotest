package auth

import (
	"errors"
	"time"
)

const (
	RoleUser          int16 = 1
	RoleModerator     int16 = 2
	RoleAdministrator int16 = 3
)

func RoleName(role int16) string {
	switch role {
	case RoleModerator:
		return "moderator"
	case RoleAdministrator:
		return "administrator"
	default:
		return "user"
	}
}

var ErrAccountBanned = errors.New("account banned")

type AccountBannedError struct {
	Reason    string
	ExpiresAt *time.Time
}

func (e *AccountBannedError) Error() string { return ErrAccountBanned.Error() }
func (e *AccountBannedError) Unwrap() error { return ErrAccountBanned }
