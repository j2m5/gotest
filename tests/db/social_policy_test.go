package db_test

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/social"
	"testing"
	"time"
)

func socialActors(t *testing.T) (*testStorage, context.Context, *db.SocialRepository, social.Actor, social.Actor) {
	t.Helper()
	s, ctx := characterTestStorage(t, true)
	for _, v := range []struct {
		id    int
		token string
	}{{1, "one"}, {2, "two"}} {
		if err := createStoredSession(t, s, ctx, v.id, v.token, time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1; UPDATE users SET role_id=2 WHERE id=2`)
	return s, ctx, db.NewSocialRepository(s.pool), social.Actor{UserID: 1, SessionHash: "one"}, social.Actor{UserID: 2, SessionHash: "two"}
}
func TestSocialScopeHierarchySnapshotAndPaging(t *testing.T) {
	s, ctx, r, user, mod := socialActors(t)
	h := 1
	if _, err := r.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "account", DurationHours: &h, Reason: "reason"}); !errors.Is(err, social.ErrForbidden) {
		t.Fatalf("moderator account=%v", err)
	}
	if _, err := r.Ban(ctx, mod, social.BanInput{UserID: 2, Scope: "chat", DurationHours: &h, Reason: "reason"}); !errors.Is(err, social.ErrForbidden) {
		t.Fatalf("self ban=%v", err)
	}
	if _, err := r.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "forum", DurationHours: &h, Reason: "reason"}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "forum", DurationHours: &h, Reason: "reason"}); !errors.Is(err, social.ErrBanActive) {
		t.Fatalf("duplicate=%v", err)
	}
	m, err := r.Send(ctx, user, social.SendInput{Body: "snapshot", CharacterID: 1})
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET character_id=2 WHERE id=1; UPDATE characters SET name='Changed name' WHERE id=1`)
	if _, err := r.Send(ctx, user, social.SendInput{Body: "stale", CharacterID: 1}); !errors.Is(err, social.ErrCharacterChanged) {
		t.Fatalf("stale=%v", err)
	}
	if _, err := r.Send(ctx, user, social.SendInput{Body: "switch", CharacterID: 2}); !errors.Is(err, social.ErrTooFast) {
		t.Fatalf("account throttle=%v", err)
	}
	mustExec(t, ctx, s, `INSERT INTO chat_messages(user_id,character_id,author_name,author_role,body) SELECT 1,1,'snapshot','user','message' FROM generate_series(1,50)`)
	p, err := r.Chat(ctx, user, 0)
	if err != nil || len(p.Messages) != 50 || p.NextBeforeID == nil {
		t.Fatalf("page=%+v %v", p, err)
	}
	p, err = r.Chat(ctx, user, *p.NextBeforeID)
	if err != nil || len(p.Messages) != 1 || p.Messages[0].ID != m.ID || p.Messages[0].AuthorName != "First character" || p.NextBeforeID != nil {
		t.Fatalf("second=%+v %v", p, err)
	}
	mustExec(t, ctx, s, `UPDATE users SET role_id=2 WHERE id=1`)
	if _, err = r.Ban(ctx, mod, social.BanInput{UserID: 1, Scope: "chat", DurationHours: &h, Reason: "reason"}); !errors.Is(err, social.ErrForbidden) {
		t.Fatalf("current hierarchy=%v", err)
	}
	if _, err = r.Chat(ctx, social.Actor{UserID: 1, SessionHash: "two"}, 0); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("foreign token=%v", err)
	}
}
func TestSocialIndependentBanBlocksLiveSessionAndValidLogin(t *testing.T) {
	s, ctx, _, _, _ := socialActors(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("valid-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET password=$1 WHERE id=1`, string(hash))
	mustExec(t, ctx, s, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','account reason',2)`)
	if _, err = s.FindSession(ctx, "one"); !errors.Is(err, auth.ErrNotFound) {
		t.Fatalf("independent ban session=%v", err)
	}
	service := auth.NewService(s.AuthRepository, auth.Options{})
	if _, err = service.Login(ctx, auth.LoginInput{Email: "first@example.test", Password: "wrong"}); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("wrong password leaked=%v", err)
	}
	if _, err = service.Login(ctx, auth.LoginInput{Email: "first@example.test", Password: "valid-password"}); !errors.Is(err, auth.ErrAccountBanned) {
		t.Fatalf("correct password=%v", err)
	}
	var detail *auth.AccountBannedError
	if !errors.As(err, &detail) || detail.Reason != "account reason" || detail.ExpiresAt != nil {
		t.Fatalf("detail=%v", err)
	}
}
func TestSocialQueuedSendAndLoginRecheckAfterBan(t *testing.T) {
	for _, op := range []string{"send", "login"} {
		t.Run(op, func(t *testing.T) {
			s, ctx, r, user, _ := socialActors(t)
			tx, err := s.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			var pid int
			if err = tx.QueryRow(ctx, `SELECT pg_backend_pid() FROM users WHERE id=1 FOR UPDATE`).Scan(&pid); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() {
				if op == "send" {
					_, e := r.Send(ctx, user, social.SendInput{Body: "pending", CharacterID: 1})
					done <- e
				} else {
					done <- s.CreateSession(ctx, auth.SessionInput{UserID: 1, PasswordHash: "unused", TokenHash: "pending", ExpiresAt: time.Now().Add(time.Hour)})
				}
			}()
			awaitResetUserLock(t, ctx, s, pid, done)
			if _, err = tx.Exec(ctx, `INSERT INTO user_bans(user_id,scope,reason,created_by_user_id) VALUES(1,'account','account ban',2); DELETE FROM sessions WHERE user_id=1; UPDATE users SET character_id=NULL WHERE id=1`); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			err = <-done
			if op == "send" && !errors.Is(err, auth.ErrUnauthenticated) {
				t.Fatalf("pending send=%v", err)
			}
			if op == "login" && !errors.Is(err, auth.ErrAccountBanned) {
				t.Fatalf("pending login=%v", err)
			}
		})
	}
}
