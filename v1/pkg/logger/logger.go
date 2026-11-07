package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Level  string
	Format string
	Output string
}

type Logger struct {
	log zerolog.Logger
}

func New(cfg Config) *Logger {
	level := parseLevel(cfg.Level)

	var output io.Writer

	switch strings.ToLower(cfg.Output) {
	case "stderr":
		output = os.Stderr
	case "stdout", "":
		output = os.Stdout
	default:
		file, err := os.OpenFile(
			cfg.Output,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0640,
		)

		if err != nil {
			output = os.Stdout
		} else {
			output = file
		}
	}

	if strings.EqualFold(cfg.Format, "console") {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(level)

	log := zerolog.New(output).
		With().
		Timestamp().
		Str("service", "fehmi-s3").
		Logger()

	return &Logger{log: log}
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) Info() *zerolog.Event {
	return l.log.Info()
}

func (l *Logger) Debug() *zerolog.Event {
	return l.log.Debug()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.log.Warn()
}

func (l *Logger) Error() *zerolog.Event {
	return l.log.Error()
}

func (l *Logger) Trace() *zerolog.Event {
	return l.log.Trace()
}

func (l *Logger) With() zerolog.Context {
	return l.log.With()
}

func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		log: l.log.With().
			Str("component", component).
			Logger(),
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		log: l.log.With().
			Str("request_id", requestID).
			Logger(),
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		log: l.log.With().
			Err(err).
			Logger(),
	}
}

func (l *Logger) Sync() error {
	return nil
}
