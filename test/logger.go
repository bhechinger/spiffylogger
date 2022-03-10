package test

import (
	"bytes"
	"context"

	"github.com/bhechinger/spiffylogger"
	"go.uber.org/zap/zapcore"
)

// TestableContext returns a context that has a byte buffer logger injected into it
// Useful for capturing test output
func TestableContext(level zapcore.Level) (context.Context, *bytes.Buffer) {
	var buf bytes.Buffer
	l := spiffylogger.NewLogger(level)
	return spiffylogger.CtxWithLogger(context.Background(), l), &buf
}

// LogCtx provides a context with the stdio logger
func LogCtx(logLevel zapcore.Level) context.Context {
	return spiffylogger.CtxWithLogger(context.Background(), spiffylogger.NewLogger(logLevel))
}
