package test

import (
	"bufio"
	"bytes"
	"context"

	"github.com/bhechinger/spiffylogger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestableContext returns a context that has a byte buffer logger injected into it
// Useful for capturing test output
func TestableContext(level zapcore.Level) (context.Context, *bytes.Buffer) {
	var b bytes.Buffer
	buf := bufio.NewWriter(&b)

	l := spiffylogger.NewLogger(level, zap.ErrorOutput(zapcore.AddSync(buf)))
	return spiffylogger.CtxWithLogger(context.Background(), l), &b
}

// LogCtx provides a context with the stdio logger
func LogCtx(logLevel zapcore.Level) context.Context {
	return spiffylogger.CtxWithLogger(context.Background(), spiffylogger.NewLogger(logLevel))
}
