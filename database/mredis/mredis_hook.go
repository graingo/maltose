package mredis

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/graingo/maltose/os/mlog"
	redis "github.com/redis/go-redis/v9"
)

// --- loggerHook ---

// loggerHook is a redis hook for logging.
type loggerHook struct {
	logger        *mlog.Logger
	slowThreshold time.Duration
	mu            sync.RWMutex
}

// newLoggerHook creates a new logger hook.
func newLoggerHook(cfg *Config) *loggerHook {
	return &loggerHook{
		logger:        cfg.Logger,
		slowThreshold: cfg.SlowThreshold,
	}
}

// setSlowThreshold updates the slow threshold for the hook.
func (h *loggerHook) setSlowThreshold(d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.slowThreshold = d
}

// DialHook is called when a connection is dialed. It's part of the redis.Hook interface.
func (h *loggerHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// We don't log dialing, just pass it through.
		return next(ctx, network, addr)
	}
}

// ProcessHook is called before a command is processed. It's part of the redis.Hook interface.
func (h *loggerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		cost := time.Since(start)

		fields := []mlog.Field{
			mlog.String("cmd", cmd.Name()),
			mlog.Duration("cost", cost),
		}

		if err != nil && err != redis.Nil {
			// The logger's Errorw method will automatically handle adding the error as a field.
			h.logger.Errorw(ctx, err, "redis command error", fields...)
		} else {
			h.mu.RLock()
			slow := h.slowThreshold
			h.mu.RUnlock()
			if slow > 0 && cost > slow {
				h.logger.Warnw(ctx, "redis command slow", fields...)
			} else {
				h.logger.Debugw(ctx, "redis command", fields...)
			}
		}

		return err
	}
}

// ProcessPipelineHook is called before a pipeline is processed. It's part of the redis.Hook interface.
func (h *loggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		cost := time.Since(start)

		fields := []mlog.Field{
			mlog.String("cmd", "pipeline"),
			mlog.Int("num_cmds", len(cmds)),
			mlog.Duration("cost", cost),
		}

		if err != nil && err != redis.Nil {
			// The logger's Errorw method will automatically handle adding the error as a field.
			h.logger.Errorw(ctx, err, "redis pipeline error", fields...)
		} else {
			h.mu.RLock()
			slow := h.slowThreshold
			h.mu.RUnlock()
			if slow > 0 && cost > slow {
				h.logger.Warnw(ctx, "redis pipeline slow", fields...)
			} else {
				h.logger.Debugw(ctx, "redis pipeline", fields...)
			}
		}

		return err
	}
}
