package mlogs

import "log/slog"

type Handler slog.Handler

type Hook interface {
	New(Handler) Handler
}
