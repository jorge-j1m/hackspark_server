package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ContextKey string

const (
	// RequestIDCtxKey is the context key for the request ID
	RequestIDCtxKey ContextKey = "request_id"
)

// Setup configures the global logger with appropriate settings based on environment
func Setup(logLevel, environment string) error {
	// Configure log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)

	// Set pretty logging for development
	var output io.Writer = os.Stdout
	if environment == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global logger
	log.Logger = zerolog.New(output).With().
		Timestamp().
		Caller().
		Stack().
		Logger()

	return nil
}

// LogWithContext returns a logger with request ID from context if present
func LogWithContext(ctx context.Context) *zerolog.Logger {
	logger := log.Logger
	if ctx == nil {
		return &logger
	}

	if reqID, ok := ctx.Value(RequestIDCtxKey).(string); ok && reqID != "" {
		logger = logger.With().Str("request_id", reqID).Logger()
	}

	return &logger
}

// Debug returns a debug event logger with optional context
func Debug(ctx context.Context) *zerolog.Event {
	return LogWithContext(ctx).Debug()
}

// Info returns an info event logger with optional context
func Info(ctx context.Context) *zerolog.Event {
	return LogWithContext(ctx).Info()
}

// Warn returns a warn event logger with optional context
func Warn(ctx context.Context) *zerolog.Event {
	return LogWithContext(ctx).Warn()
}

// Error returns an error event logger with optional context
func Error(ctx context.Context) *zerolog.Event {
	return LogWithContext(ctx).Error()
}

// Fatal returns a fatal event logger with optional context
func Fatal(ctx context.Context) *zerolog.Event {
	return LogWithContext(ctx).Fatal()
}
