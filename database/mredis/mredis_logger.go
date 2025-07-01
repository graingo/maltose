package mredis

import (
	"context"
	"net"
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/redis/go-redis/v9"
)

// loggerHook is a hook for go-redis that logs commands, errors, and slow queries.
type loggerHook struct {
	logger        *mlog.Logger
	slowThreshold time.Duration
}

var _ redis.Hook = (*loggerHook)(nil)

// newLoggerHook creates and returns a new LoggerHook.
func newLoggerHook(logger *mlog.Logger, slowThreshold time.Duration) *loggerHook {
	return &loggerHook{
		logger:        logger,
		slowThreshold: slowThreshold,
	}
}

// DialHook is a hook for dialing a new connection.
func (h *loggerHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook is a hook for processing a single command.
func (h *loggerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		startTime := time.Now()
		err := next(ctx, cmd)
		duration := time.Since(startTime)

		fields := mlog.Fields{
			mlog.String("command", cmd.Name()),
			mlog.Any("args", cmd.Args()),
			mlog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		}

		if err != nil && err != redis.Nil {
			h.logger.Errorw(ctx, err, "redis command error", fields...)
		} else if h.slowThreshold > 0 && duration > h.slowThreshold {
			h.logger.Warnw(ctx, "redis slow command", fields...)
		} else {
			h.logger.Infow(ctx, "redis command", fields...)
		}
		return err
	}
}

// ProcessPipelineHook is a hook for processing a pipeline of commands.
func (h *loggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		startTime := time.Now()
		err := next(ctx, cmds)
		duration := time.Since(startTime)

		var (
			cmdNames []string
			cmdArgs  [][]interface{}
		)
		for _, cmd := range cmds {
			cmdNames = append(cmdNames, cmd.Name())
			cmdArgs = append(cmdArgs, cmd.Args())
		}

		fields := mlog.Fields{
			mlog.Any("commands", cmdNames),
			mlog.Any("args", cmdArgs),
			mlog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		}

		if err != nil && err != redis.Nil {
			h.logger.Errorw(ctx, err, "redis pipeline error", fields...)
		} else if h.slowThreshold > 0 && duration > h.slowThreshold {
			h.logger.Warnw(ctx, "redis slow pipeline", fields...)
		} else {
			h.logger.Infow(ctx, "redis pipeline", fields...)
		}
		return err
	}
}
