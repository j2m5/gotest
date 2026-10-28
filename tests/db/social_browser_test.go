package db_test

import (
	"context"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gotest/internal/auth"
	"gotest/internal/character"
	"gotest/internal/db"
	"gotest/internal/handlers"
	"gotest/internal/middleware"
	"gotest/internal/realtime"
	"gotest/internal/routes"
	"gotest/internal/social"
	"gotest/tests/testenv"
)

func TestSocialBrowserFlow(t *testing.T) {
	browser := testenv.Browser(t)
	s, ctx := characterTestStorage(t, true)
	hash, err := bcrypt.GenerateFromPassword([]byte("browser-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `UPDATE users SET password=$1`, string(hash))
	mustExec(t, ctx, s, `UPDATE users SET character_id=1 WHERE id=1; UPDATE users SET character_id=3,role_id=2 WHERE id=2`)
	mustExec(t, ctx, s, `INSERT INTO users(email,login,password,email_verified_at,role_id) VALUES('admin@example.test','admin',$1,clock_timestamp(),3)`, string(hash))
	mustExec(t, ctx, s, `INSERT INTO characters(user_id,faction_id,class_id,name) SELECT 3,faction_id,id,'Administrator' FROM classes WHERE code='sith_warrior'; UPDATE users SET character_id=4 WHERE id=3`)
	lifecycle, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := auth.NewService(s.AuthRepository, auth.Options{})
	hub := realtime.New(lifecycle, a, "")
	h, err := routes.New(handlers.NewHandler(a, false, a), handlers.NewCharacterHandler(character.NewService(db.NewCharacterRepository(s.pool))), middleware.NewMiddleware(a), routes.Options{
		StaticDir: filepath.Join("..", "..", "frontend", "dist"), Lifecycle: lifecycle, Social: handlers.NewSocialHandler(social.NewService(db.NewSocialRepository(s.pool)), hub.Notify), ChatEvents: hub.ServeHTTP,
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(h)
	defer server.Close()
	browserCtx, stop := context.WithTimeout(context.Background(), 100*time.Second)
	defer stop()
	command := exec.CommandContext(browserCtx, "node", filepath.Join("..", "frontend", "social.browser.mjs"), server.URL, browser, t.TempDir())
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("social browser: %v\n%s", err, output)
	}
	t.Log(string(output))
}
