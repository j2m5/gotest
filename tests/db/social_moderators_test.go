package db_test

import (
	"errors"
	"testing"

	"gotest/internal/auth"
	"gotest/internal/social"
)

func TestChatHistoryReflectsCurrentAccountRole(t *testing.T) {
	s, ctx, r, user, _ := socialActors(t)
	message, err := r.Send(ctx, user, social.SendInput{Body: "Original message", CharacterID: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []int16{auth.RoleModerator, auth.RoleAdministrator, auth.RoleUser} {
		mustExec(t, ctx, s, `UPDATE users SET role_id=$1 WHERE id=1`, role)
		page, err := r.Chat(ctx, user, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := page.Messages[0]
		if got.AuthorRole != auth.RoleName(role) || got.AuthorName != message.AuthorName || got.CharacterID != message.CharacterID {
			t.Fatalf("current role %d: message=%+v", role, got)
		}
	}
}

func TestModeratorDirectoryIsAccountBasedAndPublicToPlayers(t *testing.T) {
	s, ctx, r, user, _ := socialActors(t)
	// Account 1 owns two characters, but appears only once after promotion.
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password,email_verified_at) VALUES('third@example.test','third','unused',clock_timestamp()); UPDATE users SET role_id=3 WHERE id=1`)
	list, err := r.Moderators(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Moderators) != 2 {
		t.Fatalf("staff=%+v", list.Moderators)
	}
	admin, mod := list.Moderators[0], list.Moderators[1]
	if admin.UserID != 1 || admin.Role != "administrator" || admin.Login != "first" || admin.CharacterName == nil || *admin.CharacterName != "First character" {
		t.Fatalf("admin=%+v", admin)
	}
	if mod.UserID != 2 || mod.Role != "moderator" || mod.CharacterName != nil {
		t.Fatalf("mod=%+v", mod)
	}
	mustExec(t, ctx, s, `UPDATE users SET role_id=1 WHERE id=1`)
	list, err = r.Moderators(ctx, user)
	if err != nil || len(list.Moderators) != 1 {
		t.Fatalf("regular reader: %+v %v", list, err)
	}
	// An unverified or account-banned member cannot currently moderate.
	mustExec(t, ctx, s, `UPDATE users SET email_verified_at=NULL WHERE id=2`)
	list, err = r.Moderators(ctx, user)
	if err != nil || len(list.Moderators) != 0 || list.Moderators == nil {
		t.Fatalf("unverified member: %+v %v", list, err)
	}
	mustExec(t, ctx, s, `UPDATE users SET email_verified_at=clock_timestamp() WHERE id=2; INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(2,'account','Account restriction',1)`)
	list, err = r.Moderators(ctx, user)
	if err != nil || len(list.Moderators) != 0 {
		t.Fatalf("banned member: %+v %v", list, err)
	}
	mustExec(t, ctx, s, `DELETE FROM sessions WHERE user_id=1`)
	if _, err = r.Moderators(ctx, user); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("revoked session=%v", err)
	}
}
