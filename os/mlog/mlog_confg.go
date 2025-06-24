package mlog

import (
	"io"
	"os"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/mconv"
	"github.com/sirupsen/logrus"
)

// loadConfig
func (l *Logger) loadConfig(config *Config) error {
	// Set log level
	if config.Level > 0 {
		if lvl, err := logrus.ParseLevel(mconv.ToString(config.Level)); err == nil {
			l.parent.SetLevel(lvl)
			l.config.Level = Level(lvl)
		}
	}

	// CtxKeys Hook. Always remove the existing hook first.
	l.RemoveHookByType(&ctxHook{})
	if len(config.CtxKeys) > 0 {
		l.AddHook(&ctxHook{keys: config.CtxKeys})
	}

	// Set output
	var outputs []io.Writer
	if config.Stdout {
		outputs = append(outputs, os.Stdout)
	}

	// Set file output
	if config.Path != "" && config.File != "" {
		fileWriter, err := newFileWriter(config.Path, config.File, config.AutoClean)
		if err != nil {
			return err
		}
		outputs = append(outputs, fileWriter)
	}

	// Set combined output
	if len(outputs) > 0 {
		if len(outputs) == 1 {
			l.parent.SetOutput(outputs[0])
		} else {
			l.parent.SetOutput(io.MultiWriter(outputs...))
		}
	} else {
		// If there is no output, set to io.Discard
		l.parent.SetOutput(io.Discard)
	}

	// Set time format
	timeFormat := config.TimeFormat
	if timeFormat == "" {
		timeFormat = defaultTimeFormat
	}

	// Set log format
	formatStr := config.Format
	switch formatStr {
	case "json":
		l.parent.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: timeFormat,
		})
	case "text", "": // Default to text format
		l.parent.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: timeFormat,
			FullTimestamp:   true,
		})
	default:
		return merror.NewCodef(mcode.CodeInvalidParameter, `invalid format: %s`, formatStr)
	}

	return nil
}

func (l *Logger) SetConfigWithMap(configMap map[string]any) error {
	if err := l.config.SetConfigWithMap(configMap); err != nil {
		return err
	}
	return l.loadConfig(l.config)
}

func (l *Logger) SetConfig(config *Config) error {
	l.config = config
	return l.loadConfig(config)
}
