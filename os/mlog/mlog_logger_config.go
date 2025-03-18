package mlog

import (
	"io"
	"os"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/mconv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Level      Level    `json:"level"`
	Path       string   `json:"path"`
	File       string   `json:"file"`
	TimeFormat string   `json:"time_format"`
	Format     string   `json:"format"`
	Stdout     bool     `json:"stdout"`
	AutoClean  int      `json:"auto_clean"`
	CtxKeys    []string `json:"ctx_keys"`
}

func DefaultConfig() Config {
	return Config{
		Level:      InfoLevel,
		Path:       defaultPath,
		File:       defaultFile,
		TimeFormat: defaultTimeFormat,
		Format:     defaultFormat,
		Stdout:     true,
		CtxKeys:    []string{},
	}
}

func (l *Logger) SetConfig(config Config) error {
	return l.SetConfigWithMap(map[string]any{
		"level":       config.Level,
		"path":        config.Path,
		"file":        config.File,
		"time_format": config.TimeFormat,
		"format":      config.Format,
		"stdout":      config.Stdout,
		"auto_clean":  config.AutoClean,
		"ctx_keys":    config.CtxKeys,
	})
}

// SetConfigWithMap sets the logger configuration using a map.
func (l *Logger) SetConfigWithMap(config map[string]any) error {
	// Set log level
	if v, ok := config["level"]; ok {
		if lvl, err := logrus.ParseLevel(mconv.ToString(v)); err == nil {
			l.parent.SetLevel(lvl)
			l.config.Level = Level(lvl)
		}
	}

	// Update config values
	if v, ok := config["path"]; ok {
		l.config.Path = mconv.ToString(v)
	}

	if v, ok := config["file"]; ok {
		l.config.File = mconv.ToString(v)
	}

	if v, ok := config["auto_clean"]; ok {
		l.config.AutoClean = mconv.ToInt(v)
	}

	if v, ok := config["ctx_keys"]; ok {
		if keys, ok := v.([]string); ok {
			l.RemoveHookByType(&ctxHook{})
			l.config.CtxKeys = keys
			l.AddHook(&ctxHook{keys: keys})
		}
	}

	// Set output
	var outputs []io.Writer
	// Set stdout
	stdout := true // default enable stdout
	if v, ok := config["stdout"]; ok {
		stdout = mconv.ToBool(v)
		l.config.Stdout = stdout
	}
	if stdout {
		outputs = append(outputs, os.Stdout)
	}

	// Set file output
	if l.config.Path != "" && l.config.File != "" {
		fileWriter, err := newFileWriter(l.config.Path, l.config.File, l.config.AutoClean)
		if err != nil {
			return err
		}
		outputs = append(outputs, fileWriter)
	}

	// Set output
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
	timeFormat := defaultTimeFormat
	if v, ok := config["time_format"]; ok {
		timeFormat = mconv.ToString(v)
		l.config.TimeFormat = timeFormat
	}

	// Set log format
	if format, ok := config["format"]; ok {
		formatStr := mconv.ToString(format)
		l.config.Format = formatStr

		switch formatStr {
		case "json":
			l.parent.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: timeFormat,
			})
		case "text":
			l.parent.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: timeFormat,
				FullTimestamp:   true,
			})
		default:
			return merror.NewCodef(mcode.CodeInvalidParameter, `invalid format: %s`, format)
		}
	}

	return nil
}

// SetStdoutPrint sets the stdout print.
func (l *Logger) SetStdoutPrint(enabled bool) {
	l.SetConfigWithMap(map[string]any{
		"stdout": enabled,
	})
}

// SetPath sets the log file path.
func (l *Logger) SetPath(path string) {
	l.SetConfigWithMap(map[string]any{
		"path": path,
	})
}

// SetFile sets the log file name.
func (l *Logger) SetFile(file string) {
	l.SetConfigWithMap(map[string]any{
		"file": file,
	})
}

// SetAutoClean sets the auto clean days.
func (l *Logger) SetAutoClean(autoClean int) {
	l.SetConfigWithMap(map[string]any{
		"auto_clean": autoClean,
	})
}

// SetTimeFormat sets the log time format.
func (l *Logger) SetTimeFormat(timeFormat string) {
	l.SetConfigWithMap(map[string]any{
		"time_format": timeFormat,
	})
}

// SetFormat sets the log format.
func (l *Logger) SetFormat(format string) {
	l.SetConfigWithMap(map[string]any{
		"format": format,
	})
}

// SetCtxKeys sets the context keys to extract.
func (l *Logger) SetCtxKeys(keys []string) {
	l.SetConfigWithMap(map[string]any{
		"ctx_keys": keys,
	})
}
