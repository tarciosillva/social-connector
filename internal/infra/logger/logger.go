package logger

import (
	"context"
	"os"
	"social-connector/internal/config"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger *logrus.Logger
	ctx    context.Context
}

// NewLogger initializes a new instance of Logger with optional configuration.
func NewLogger(ctx context.Context, jsonFormat bool) *Logger {
	logger := logrus.New()
	logger.Out = os.Stdout

	logLevel := config.GetEnv("LOG_LEVEL")

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if jsonFormat {
		logger.SetFormatter(&logrus.JSONFormatter{
			PrettyPrint: false,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
			PadLevelText:  true,
		})
	}

	return &Logger{logger: logger, ctx: ctx}
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, fields ...logrus.Fields) {
	l.logWithFields(logrus.DebugLevel, msg, fields...)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string, fields ...logrus.Fields) {
	l.logWithFields(logrus.InfoLevel, msg, fields...)
}

// Warn logs a warn-level message.
func (l *Logger) Warn(msg string, fields ...logrus.Fields) {
	l.logWithFields(logrus.WarnLevel, msg, fields...)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, fields ...logrus.Fields) {
	l.logWithFields(logrus.ErrorLevel, msg, fields...)
}

// Fatal logs a fatal-level message and exits the application.
func (l *Logger) Fatal(msg string, fields ...logrus.Fields) {
	l.logWithFields(logrus.FatalLevel, msg, fields...)
	os.Exit(1) // Ensure application exits after fatal log
}

// logWithFields logs a message with optional fields at the specified level.
func (l *Logger) logWithFields(level logrus.Level, msg string, fields ...logrus.Fields) {
	entry := l.logger.WithContext(l.ctx)

	if len(fields) > 0 {
		for _, field := range fields {
			entry = entry.WithFields(field)
		}
	}

	entry.Log(level, msg)
}
