package db_test

import (
	"errors"
	"gotest/internal/auth"
	"gotest/internal/social"
	"testing"
)

func TestSocialQueuedModerationRechecksRoleAndSession(t *testing.T) {
	for _, change := range []string{"role", "session"} {
		t.Run(change, func(t *testing.T) {
			s, ctx, r, _, mod := socialActors(t)
			tx, pid := lockResetUser(t, ctx, s)
			defer tx.Rollback(ctx)
			done := make(chan error, 1)
			h := 1
			go func() {
				_, err := r.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "chat", DurationHours: &h, Reason: "reason"})
				done <- err
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			sql := `UPDATE users SET role_id=1 WHERE id=2`
			if change == "session" {
				sql = `DELETE FROM sessions WHERE user_id=2`
			}
			if _, err := tx.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			err := <-done
			expected := social.ErrForbidden
			if change == "session" {
				expected = auth.ErrUnauthenticated
			}
			if !errors.Is(err, expected) {
				t.Fatalf("stale %s=%v", change, err)
			}
		})
	}
}

func TestSocialAccountBanRollsBackWhenJournalFails(t *testing.T) {
	s, ctx, r, user, admin := socialActors(t)
	mustExec(t, ctx, s, `UPDATE users SET role_id=3 WHERE id=2; ALTER TABLE moderation_actions ADD CONSTRAINT reject_audit CHECK (false)`)
	if _, err := r.Ban(ctx, admin, social.BanInput{UserID: 1, Scope: "account", Reason: "reason"}); err == nil {
		t.Fatal("ban ignored audit failure")
	}
	if _, err := s.FindSession(ctx, user.SessionHash); err != nil {
		t.Fatalf("failed ban revoked session: %v", err)
	}
	if activeCharacter(t, ctx, s, 1) == nil {
		t.Fatal("failed ban cleared character")
	}
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM user_bans`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial ban count=%d %v", count, err)
	}
}

func TestSocialExpiryAndRevocationUseDatabaseTime(t *testing.T) {
	s, ctx, r, user, mod := socialActors(t)
	mustExec(t, ctx, s, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id,created_at,expires_at) VALUES(1,'chat','old ban',2,now()-interval '2 hours',now()-interval '1 hour')`)
	page, err := r.Chat(ctx, user, 0)
	if err != nil || page.ChatBan != nil {
		t.Fatalf("expired ban=%+v %v", page, err)
	}
	h := 1
	if _, err = r.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "chat", DurationHours: &h, Reason: "new ban"}); err != nil {
		t.Fatalf("expired ban conflict=%v", err)
	}
	if _, err = r.Send(ctx, user, social.SendInput{Body: "blocked", CharacterID: 1}); !errors.Is(err, social.ErrChatBanned) {
		t.Fatalf("active new ban=%v", err)
	}
}
