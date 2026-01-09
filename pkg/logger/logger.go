package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

func New(level string) *Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logger := zerolog.New(output).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger}
}

func (l *Logger) WithComponent(name string) *Logger {
	return &Logger{l.With().Str("component", name).Logger()}
}
