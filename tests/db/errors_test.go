package db_test

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgxpool"
	"gotest/internal/auth"
	"gotest/internal/db"
)

// Exercise error conversion through the public repository and a real pgx
// connection. This protocol stub supplies responses, not SQL semantics;
// migration/transaction checks still require TEST_DATABASE_URL.
func protocolStorage(t *testing.T, code string) *db.AuthRepository {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	t.Cleanup(func() { listener.Close(); <-done })
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				t.Error(err)
			}
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		backend := pgproto3.NewBackend(conn, conn)
		if _, err = backend.ReceiveStartupMessage(); err != nil {
			t.Error(err)
			return
		}
		backend.Send(&pgproto3.AuthenticationOk{})
		backend.Send(&pgproto3.ParameterStatus{Name: "client_encoding", Value: "UTF8"})
		backend.Send(&pgproto3.ParameterStatus{Name: "standard_conforming_strings", Value: "on"})
		backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
		if err = backend.Flush(); err != nil {
			t.Error(err)
			return
		}
		txStatus := byte('I')
		for {
			message, err := backend.Receive()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					t.Error(err)
				}
				return
			}
			switch message := message.(type) {
			case *pgproto3.Terminate:
				return
			case *pgproto3.Query:
				// Fail the repository query, not BEGIN: pgx deliberately destroys a
				// connection after a failed BEGIN, which races the stub's next read.
				switch strings.ToLower(strings.TrimSpace(message.String)) {
				case "begin":
					txStatus = 'T'
					backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("BEGIN")})
				case "rollback":
					txStatus = 'I'
					backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("ROLLBACK")})
				default:
					if code == "" {
						backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("SELECT 0")})
					} else {
						backend.Send(&pgproto3.ErrorResponse{Severity: "ERROR", Code: code, Message: "test database response"})
						if txStatus == 'T' {
							txStatus = 'E'
						}
					}
				}
				backend.Send(&pgproto3.ReadyForQuery{TxStatus: txStatus})
				if err = backend.Flush(); err != nil {
					t.Error(err)
					return
				}
			default:
				t.Errorf("unexpected protocol message: %T", message)
				return
			}
		}
	}()
	config, err := pgxpool.ParseConfig("postgres://test@" + listener.Addr().String() + "/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return db.NewAuthRepository(pool)
}

func TestAuthErrorClassification(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		want       error
	}{
		{"missing row", "", auth.ErrNotFound},
		{"unique constraint", "23505", auth.ErrConflict},
		{"cancelled", "", context.Canceled},
		{"other database error", "23503", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := protocolStorage(t, tc.code)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var err error
			if tc.name == "missing row" {
				_, err = s.FindCredentialsByEmail(ctx, "missing@example.test")
			} else {
				if tc.name == "cancelled" {
					cancel()
				}
				err = s.CreateSession(ctx, auth.SessionInput{UserID: 1, TokenHash: "token", ExpiresAt: time.Now().Add(time.Hour)})
			}
			if tc.want != nil {
				if !errors.Is(err, tc.want) {
					t.Fatalf("got %v, want %v", err, tc.want)
				}
				return
			}
			var postgres *pgconn.PgError
			if !errors.As(err, &postgres) || postgres.Code != "23503" {
				t.Fatalf("unexpected conversion: %v", err)
			}
		})
	}
}
