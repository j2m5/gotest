package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
)

type requestContext struct {
	id     string
	logger *slog.Logger
}
type requestKey struct{}

func WithRequest(ctx context.Context, logger *slog.Logger, id string) context.Context {
	return context.WithValue(ctx, requestKey{}, requestContext{id, logger.With("request_id", id)})
}

func Logger(ctx context.Context) *slog.Logger {
	if request, ok := ctx.Value(requestKey{}).(requestContext); ok {
		return request.logger
	}
	return slog.Default()
}

func RequestID(ctx context.Context) string {
	request, _ := ctx.Value(requestKey{}).(requestContext)
	return request.id
}

// ErrorAttrs deliberately excludes Error(): driver and SMTP errors can contain
// credentials, SQL values or message recipients. Keep safe diagnostic categories.
func ErrorAttrs(err error) []slog.Attr {
	kind := "internal"
	var state interface{ SQLState() string }
	var network net.Error
	attrs := []slog.Attr{slog.String("error_type", fmt.Sprintf("%T", err))}
	switch {
	case errors.Is(err, context.Canceled):
		kind = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		kind = "timeout"
	case errors.As(err, &state):
		kind = "database"
		code := state.SQLState()
		valid := len(code) == 5
		for _, c := range code {
			valid = valid && (c >= '0' && c <= '9' || c >= 'A' && c <= 'Z')
		}
		if valid {
			attrs = append(attrs, slog.String("sqlstate", code))
		}
	case errors.As(err, &network):
		kind = "network"
		if network.Timeout() {
			kind = "timeout"
		}
	}
	return append(attrs, slog.String("error_kind", kind))
}
