package db_test

import (
	"errors"
	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/social"
	"testing"
	"time"
)

func TestSocialChatModerationAndAccountBan(t *testing.T) {
	s, ctx := characterTestStorage(t, true)
	r := db.NewSocialRepository(s.pool)
	user := social.Actor{UserID: 1, SessionHash: "user"}
	admin := social.Actor{UserID: 2, SessionHash: "admin"}
	if err := createStoredSession(t, s, ctx, 1, "user", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := createStoredSession(t, s, ctx, 2, "admin", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET role_id=3 WHERE id=2; UPDATE users SET character_id=1 WHERE id=1`)
	msg, err := r.Send(ctx, user, social.SendInput{Body: "hello", CharacterID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.Send(ctx, user, social.SendInput{Body: "too fast", CharacterID: 1}); !errors.Is(err, social.ErrTooFast) {
		t.Fatalf("throttle=%v", err)
	}
	if err = r.Delete(ctx, admin, social.DeleteInput{MessageID: msg.ID, Reason: "reason"}); err != nil {
		t.Fatal(err)
	}
	page, err := r.Chat(ctx, user, 0)
	if err != nil || len(page.Messages) != 1 || page.Messages[0].Body != nil {
		t.Fatalf("redaction=%+v %v", page, err)
	}
	h := 1
	b, err := r.Ban(ctx, admin, social.BanInput{UserID: 1, Scope: "chat", DurationHours: &h, Reason: "reason"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.Send(ctx, user, social.SendInput{Body: "blocked", CharacterID: 1}); !errors.Is(err, social.ErrChatBanned) {
		t.Fatalf("chat ban=%v", err)
	}
	if err = r.Revoke(ctx, admin, social.RevokeInput{BanID: b.ID, Reason: "lifted"}); err != nil {
		t.Fatal(err)
	}
	if _, err = r.Ban(ctx, admin, social.BanInput{UserID: 1, Scope: "account", Reason: "permanent"}); err != nil {
		t.Fatal(err)
	}
	if _, err = r.Chat(ctx, user, 0); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("revoked=%v", err)
	}
	if err = createStoredSession(t, s, ctx, 1, "new", time.Now().Add(time.Hour)); !errors.Is(err, auth.ErrAccountBanned) {
		t.Fatalf("login blocked=%v", err)
	}
	if activeCharacter(t, ctx, s, 1) != nil {
		t.Fatal("account ban retained selection")
	}
	journal, err := r.Actions(ctx, admin, 0)
	if err != nil || len(journal.Actions) != 4 {
		t.Fatalf("audit=%+v %v", journal, err)
	}
	bans, err := r.Bans(ctx, admin, 0, 1)
	if err != nil || len(bans.Bans) != 2 || bans.Bans[1].RevokedAt == nil {
		t.Fatalf("ban history=%+v %v", bans, err)
	}
	var body string
	if err = s.pool.QueryRow(ctx, `SELECT body FROM chat_messages WHERE id=$1`, msg.ID).Scan(&body); err != nil || body != "hello" {
		t.Fatalf("stored deleted body=%q %v", body, err)
	}
}
