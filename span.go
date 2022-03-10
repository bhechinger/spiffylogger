package spiffylogger

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/go-stack/stack"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap/zapcore"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type (
	spanKey struct{}
)

// Span is our implementation of a Spanner
type Span struct {
	name  string
	start time.Time
	cID   string
	sID   string
	ll    *LeveledLogger
}

// OpenSpan configures and returns a Span from a context, creating a child span if one exists in the current context
func OpenSpan(ctx context.Context) (context.Context, *Span) {
	caller := "unknown"
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		d := runtime.FuncForPC(pc)
		if d != nil {
			n := strings.Split(d.Name(), "/")
			caller = n[len(n)-1] // get just the filename + function for our span's name
		}
	}
	return openNamedSpan(ctx, caller, 1)
}

// OpenCustomSpan configures and returns a Span from a context, creating a child span if one exists in the current context
// "custom" only if we want a custom name for this span
func OpenCustomSpan(ctx context.Context, name string) (context.Context, *Span) {
	return openNamedSpan(ctx, name, 1)
}

// openNamedSpan contains the common code for OpenSpan and OpenCustomSpan
// with the appropriate log depth of 3
func openNamedSpan(ctx context.Context, name string, depth int) (context.Context, *Span) {
	depth++
	var newSpan *Span
	if s, ok := spanFromContext(ctx); ok {
		newSpan = openChildSpan(s, name, depth)
	} else {
		l, ok := loggerFromContext(ctx)
		if !ok {
			// if we don't get a logger, make sure we're at least logging to stderr
			l = NewLogger(zapcore.InfoLevel)
		}
		newSpan = openNewSpan(name, l, depth)
		if !ok {
			newSpan.printToLog(zapcore.InfoLevel, "failed to find logger in context; defaulting to stderr logger", depth)
		}
	}
	return CtxWithSpan(ctx, newSpan), newSpan
}

// openNew returns a child span of this span, keeping the same context and CID
func openChildSpan(s *Span, childName string, depth int) *Span {
	depth++
	ns := &Span{
		name:  fmt.Sprintf("%s|%s", s.name, childName), // semi-stacktrace naming
		start: time.Now(),
		cID:   s.cID,
		ll:    s.ll,
	}
	ns.sID = ns.newID(depth)

	if s.ll.Level >= zapcore.DebugLevel {
		ns.printToLog(zapcore.DebugLevel, "span opened (child)", depth)
	}
	return ns
}

// spanFromContext pulls a span out of a context
func spanFromContext(ctx context.Context) (s *Span, found bool) {
	key, ok := ctx.Value(spanKey{}).(*Span)
	if ok {
		return key, true
	}
	return nil, false
}

// CtxWithSpan allows for injecting a span into a context
func CtxWithSpan(ctx context.Context, s *Span) context.Context {
	return context.WithValue(ctx, spanKey{}, s)
}

// openNew returns a brand new span with a new CID
func openNewSpan(name string, l *LeveledLogger, depth int) *Span {
	depth++
	s := &Span{
		name:  name,
		start: time.Now(),
		ll:    l,
	}
	s.cID = s.newID(depth)
	s.sID = s.newID(depth)
	if s.ll.Level >= zapcore.DebugLevel {
		s.printToLog(zapcore.DebugLevel, "span opened", 1)
	}
	return s
}

func (s *Span) newID(depth int) string {
	depth++
	id, err := ksuid.NewRandom()
	if err != nil {
		s.printToLog(zapcore.ErrorLevel, errors.Wrap(err, "Failed to generate id.").Error(), depth)
		return "ERRID"
	}
	return id.String()
}

// Close .
func (s *Span) Close() {
	// TODO MONSTRO-749: close/end OT span
	// TODO MONSTRO-754: add timing metric to OT
	dur := time.Since(s.start)
	if s.ll.Level >= zapcore.DebugLevel {
		s.printToLog(zapcore.DebugLevel, fmt.Sprintf("span closed dur=%dns", dur), 1)
	}
}

// Error .
func (s *Span) Error(err error) {
	if s.ll.Level >= zapcore.ErrorLevel {
		s.printToLog(zapcore.ErrorLevel, withStacktrace(err), 1)
	}
}

// Errorf .
func (s *Span) Errorf(err error, fs string, v ...interface{}) {
	if s.ll.Level >= zapcore.ErrorLevel {
		message := fmt.Sprintf(fs, v...)
		s.printToLog(zapcore.ErrorLevel, fmt.Sprintf("%s: %s", message, withStacktrace(err)), 1)
	}
}

func withStacktrace(err error) string {
	// %+v gives us the error message plus a full stack trace for the error, as long as it was constructed with the "github.com/pkg/errors" package
	// we should strive to use `errors.New`, `errors.Errorf`, and `errors.Wrap` wherever we create a new error or get one from an external source
	return fmt.Sprintf("%+v", err)
}

// Info .
func (s *Span) Info(msg string) {
	if s.ll.Level >= zapcore.InfoLevel {
		s.printToLog(zapcore.InfoLevel, msg, 1)
	}
}

// Infof .
func (s *Span) Infof(fs string, v ...interface{}) {
	if s.ll.Level >= zapcore.InfoLevel {
		s.printToLog(zapcore.InfoLevel, fmt.Sprintf(fs, v...), 1)
	}
}

// Debug .
func (s *Span) Debug(msg string) {
	if s.ll.Level >= zapcore.DebugLevel {
		s.printToLog(zapcore.DebugLevel, msg, 1)
	}
}

// Debugf .
func (s *Span) Debugf(fs string, v ...interface{}) {
	if s.ll.Level >= zapcore.DebugLevel {
		s.printToLog(zapcore.DebugLevel, fmt.Sprintf(fs, v...), 1)
	}
}

// printToLog is solely responsible for creating log lines and printing them to the logger
//
// NOTE about log levels: we want to check levels before calling this function
// to avoid string cacentation functions being called needlessly
//
// NOTE: we want to use printToLog explicitly in our logging functions to ensure the caller is captured correctly (exactly 2 function callers away)
//
// NOTE: depth is relative to the calls in this package. We always want depth to be equal to the call of these functions.
// Therefore, its important to be careful to not call spans's public-facing functions inside of span.
// Instead, each internal function should accept a depth value, and +1 that value for its own call.
func (s *Span) printToLog(level zapcore.Level, msg string, depth int) {
	depth++
	c := stack.Caller(depth)
	n := NewLine(level, s, msg, &c)
	switch s.ll.Level {
	case zapcore.ErrorLevel:
		s.ll.Logger.Error(msg, n.Fields...)
	case zapcore.WarnLevel:
		s.ll.Logger.Warn(msg, n.Fields...)
	case zapcore.InfoLevel:
		s.ll.Logger.Info(msg, n.Fields...)
	case zapcore.DebugLevel:
		s.ll.Logger.Debug(msg, n.Fields...)
	}
}

// implement migration logging interface //TODO is there something else we can do here?

// Printf .
func (s *Span) Printf(msg string, v ...interface{}) {
	if s.ll.Level >= zapcore.DebugLevel {
		s.printToLog(zapcore.DebugLevel, fmt.Sprintf(msg, v...), 1)
	}
}

// Verbose returns true if we are at DEBUG level logging
func (s *Span) Verbose() bool {
	return s.ll.Level >= zapcore.DebugLevel
}
