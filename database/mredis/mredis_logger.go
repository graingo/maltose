package mredis

import (
	"context"
	"fmt"
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

		if err != nil && err != redis.Nil {
			h.logger.Errorf(
				ctx,
				`[MREDIS] command:"%s", args:%v, error:"%v"`,
				cmd.Name(), cmd.Args(), err,
			)
		} else if h.slowThreshold > 0 && duration > h.slowThreshold {
			h.logger.Warnf(
				ctx,
				`[MREDIS] slow command, command:"%s", args:%v, duration:"%s"`,
				cmd.Name(), cmd.Args(), duration,
			)
		} else {
			h.logger.Infof(
				ctx,
				`[MREDIS] command:"%s", args:%v, duration:"%s"`,
				cmd.Name(), cmd.Args(), duration,
			)
		}
		return err
	}
}

// ProcessPipelineHook is a hook for processing a pipeline of commands.
func (h *loggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		const pipelinePlaceholder = "pipeline"
		startTime := time.Now()
		err := next(ctx, cmds)
		duration := time.Since(startTime)

		var (
			cmdArgs [][]interface{}
		)
		for _, cmd := range cmds {
			cmdArgs = append(cmdArgs, cmd.Args())
		}
		argsStr := fmt.Sprintf("%v", cmdArgs)

		if err != nil && err != redis.Nil {
			h.logger.Errorf(
				ctx,
				`[MREDIS] command:"%s", args:%s, error:"%v"`,
				pipelinePlaceholder, argsStr, err,
			)
		} else if h.slowThreshold > 0 && duration > h.slowThreshold {
			h.logger.Warnf(
				ctx,
				`[MREDIS] slow command, command:"%s", args:%s, duration:"%s"`,
				pipelinePlaceholder, argsStr, duration,
			)
		} else {
			h.logger.Infof(
				ctx,
				`[MREDIS] command:"%s", args:%s, duration:"%s"`,
				pipelinePlaceholder, argsStr, duration,
			)
		}
		return err
	}
}
