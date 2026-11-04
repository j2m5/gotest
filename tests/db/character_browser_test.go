package db_test

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"gotest/internal/auth"
	"gotest/internal/db"
	"gotest/internal/mailer"
	"gotest/internal/notification"
	"gotest/internal/wallet"
	"gotest/tests/testenv"
)

type browserMail struct {
	flowMail
	sender notification.Sender
}

func (m *browserMail) Deliver(ctx context.Context, delivery notification.Delivery) error {
	if err := m.sender.Deliver(ctx, delivery); err != nil {
		return err
	}
	return m.flowMail.Deliver(ctx, delivery)
}

// This optional smoke test uses the real built SPA, HTTP API and isolated SQL
// schema. Only the Chromium binary is external; no browser package is required.
func TestCharacterBrowserFlow(t *testing.T) {
	browser := testenv.Browser(t)
	if _, err := os.Stat(filepath.Join("..", "..", "frontend", "dist", "index.html")); err != nil {
		t.Fatal("build frontend before the browser smoke test")
	}
	s, ctx := characterTestStorage(t, true)
	server := httptest.NewUnstartedServer(nil)
	defer server.Close()
	mailDirectory := t.TempDir()
	delivery, err := mailer.New(mailer.Options{Driver: "local", Directory: mailDirectory, PublicURL: "http://" + server.Listener.Addr().String(), From: "test@example.test"})
	if err != nil {
		t.Fatal(err)
	}
	mail := &browserMail{sender: delivery}
	h, service := characterFlowRouter(t, s, mail)
	if err := service.Register(ctx, auth.RegisterInput{Email: "browser@example.test", Login: "browser", Password: "browser-password"}); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyEmail(ctx, mail.waitVerification(t)); err != nil {
		t.Fatal(err)
	}
	if err := service.Register(ctx, auth.RegisterInput{Email: "other-browser@example.test", Login: "other-browser", Password: "browser-password"}); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyEmail(ctx, mail.waitVerification(t)); err != nil {
		t.Fatal(err)
	}
	// This account exercises valid resource boundaries and a stale local avatar.
	if err := service.Register(ctx, auth.RegisterInput{Email: "profile-fixture@example.test", Login: "profile-fixture", Password: "browser-password"}); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyEmail(ctx, mail.waitVerification(t)); err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, s, `INSERT INTO characters(user_id,faction_id,class_id,name,hp,hp_current,mp,mp_current,avatar)
		SELECT u.id,c.faction_id,c.id,'Проверка',200,37,0,0,'images/avatars/browser-missing.png'
		FROM users u CROSS JOIN classes c WHERE u.email='profile-fixture@example.test' AND c.code='imperial_agent'`)
	if err := service.Register(ctx, auth.RegisterInput{Email: "unverified@example.test", Login: "unverified", Password: "browser-password"}); err != nil {
		t.Fatal(err)
	}
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failure := r.Header.Get("X-Browser-Test-Wallet-Failure")
		if (r.URL.Path == "/api/wallet" && failure == "balance") || (r.URL.Path == "/api/wallet/history" && failure == "history") {
			http.Error(w, "wallet temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		// Test-only setup: exercise the actual server operation without adding a
		// production HTTP grant endpoint or changing application database balances.
		if r.URL.Path == "/api/wallet" && r.Header.Get("X-Browser-Test-Wallet-Setup") == "1" {
			id, err := strconv.Atoi(r.URL.Query().Get("character_id"))
			if err != nil {
				http.Error(w, "bad test character", 400)
				return
			}
			operations := []wallet.Operation{walletOperation(id, wallet.Credits, math.MaxInt64-22, "browser:initial")}
			for i := range 22 {
				operations = append(operations, walletOperation(id, wallet.Credits, 1, fmt.Sprintf("browser:grant:%d", i)))
			}
			operations = append(operations, walletOperation(id, wallet.Credits, -2, "browser:spend"), walletOperation(id, wallet.CartelCoins, 37, "browser:coins"))
			for _, operation := range operations {
				if _, err = db.NewWalletRepository(s.pool).Apply(r.Context(), operation); err != nil {
					http.Error(w, "wallet fixture failed", 500)
					return
				}
			}
		}
		// Populate/restore ownership only in this browser test's isolated schema.
		// Neutral content is exercised through the real routes without opening a
		// production control endpoint or changing the application world's owners.
		control := r.Header.Get("X-Browser-Test-Neutral-Control")
		if r.URL.Path == "/api/world/navigation" && (control == "claim" || control == "release") {
			var controller *int
			if control == "claim" {
				id, err := strconv.Atoi(r.URL.Query().Get("character_id"))
				if err != nil {
					http.Error(w, "bad test character", http.StatusBadRequest)
					return
				}
				var factionID int
				if err := s.pool.QueryRow(r.Context(), `SELECT faction_id FROM characters WHERE id=$1`, id).Scan(&factionID); err != nil {
					http.Error(w, "test faction unavailable", http.StatusInternalServerError)
					return
				}
				controller = &factionID
			}
			if _, err := s.pool.Exec(r.Context(), `UPDATE planets SET controlling_faction_id=$1 WHERE code=ANY($2) AND core_faction_id IS NULL`, controller, neutralContentPlanets); err != nil {
				http.Error(w, "neutral world fixture failed", http.StatusInternalServerError)
				return
			}
		}
		if r.URL.Path == "/api/npcs" && r.Header.Get("X-Browser-Test-NPC-Failure") == "1" {
			http.Error(w, "NPCs temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		// Exercise the empty UI independently of which worlds already have content.
		if r.URL.Path == "/api/npcs" && r.Header.Get("X-Browser-Test-NPC-Empty") == "1" {
			id, err := strconv.Atoi(r.URL.Query().Get("location_id"))
			if err != nil {
				http.Error(w, "bad test location", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"location_id":%d,"npcs":[]}`, id)
			return
		}
		if r.URL.Path == "/api/world/locations" && r.Header.Get("X-Browser-Test-Locations-Failure") == "1" {
			http.Error(w, "locations temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path == "/api/world/planets" && r.Header.Get("X-Browser-Test-World-Failure") == "1" {
			http.Error(w, "world temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path == "/api/auth/verify" && r.Header.Get("X-Browser-Test-Verification-Failure") == "1" {
			http.Error(w, "verification temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path == "/api/characters/catalog" && r.Header.Get("X-Browser-Test-Catalog-Failure") == "1" {
			http.Error(w, "catalog temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		// Chromium forbids overriding Sec-Fetch-Site through extra headers.
		// Inject that metadata at the test boundary so the real origin guard
		// produces a readable 403 for the login UI, without disabling CORS.
		if r.URL.Path == "/api/auth/login" && r.Header.Get("X-Browser-Test-Cross-Site") == "1" {
			r.Header.Set("Sec-Fetch-Site", "cross-site")
			r.Header.Set("Origin", "https://untrusted.example")
		}
		h.ServeHTTP(w, r)
	})
	server.Start()
	browserCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	command := exec.CommandContext(browserCtx, "node", filepath.Join("..", "frontend", "characters.browser.mjs"), server.URL, browser, t.TempDir(), mailDirectory)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("browser smoke test: %v\n%s", err, output)
	}
	t.Log(string(output))
}
