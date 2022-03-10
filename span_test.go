package spiffylogger_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/bhechinger/spiffylogger"
	"github.com/bhechinger/spiffylogger/test"
)

func TestSpan_Debug(t *testing.T) {
	logged := RunTestProgram(zapcore.DebugLevel)

	require.Greater(t, len(logged), 0)
	require.Contains(t, logged, "rootSpan")
	require.Contains(t, logged, "testDebug")
	require.Contains(t, logged, "testDebugf")
	require.Contains(t, logged, "testInfo")
	require.Contains(t, logged, "testInfof")
	require.Contains(t, logged, "testError")
	require.Contains(t, logged, "testErrorf")
	require.Contains(t, logged, "testGoErr")
	require.Contains(t, logged, "rootSpan|main_test.RunTestProgram")
	require.Contains(t, logged, "hello!")
	require.Contains(t, logged, "span_test.go")
	require.Contains(t, logged, "opened")
	require.Contains(t, logged, "code.chipmunq.com") // stk trace

	verifyAllJSON(t, logged)
}

func TestSpan_Info(t *testing.T) {
	logged := RunTestProgram(zapcore.InfoLevel)

	require.Greater(t, len(logged), 0)
	require.Contains(t, logged, "rootSpan")
	require.NotContains(t, logged, "testDebug")
	require.NotContains(t, logged, "testDebugf")
	require.Contains(t, logged, "testInfo")
	require.Contains(t, logged, "testInfof")
	require.Contains(t, logged, "testError")
	require.Contains(t, logged, "testErrorf")
	require.Contains(t, logged, "testGoErr")
	require.Contains(t, logged, "rootSpan|main_test.RunTestProgram")
	require.Contains(t, logged, "hello!")
	require.Contains(t, logged, "span_test.go")
	require.NotContains(t, logged, "opened")
	require.Contains(t, logged, "code.chipmunq.com") // stk trace

	verifyAllJSON(t, logged)
}

func TestSpan_Error(t *testing.T) {
	logged := RunTestProgram(zapcore.ErrorLevel)

	require.Greater(t, len(logged), 0)
	require.Contains(t, logged, "rootSpan")
	require.NotContains(t, logged, "testDebug")
	require.NotContains(t, logged, "testDebugf")
	require.NotContains(t, logged, "testInfo")
	require.NotContains(t, logged, "testInfof")
	require.Contains(t, logged, "testError")
	require.Contains(t, logged, "testErrorf")
	require.Contains(t, logged, "testGoErr")
	require.NotContains(t, logged, "rootSpan|main_test.RunTestProgram") // the only message in the child was an info level
	require.NotContains(t, logged, "hello!")
	require.Contains(t, logged, "span_test.go")
	require.NotContains(t, logged, "opened")
	require.Contains(t, logged, "code.chipmunq.com") // stk trace

	verifyAllJSON(t, logged)
}

// RunTestProgram uses a logger and prints a log to the returned string
func RunTestProgram(logLevel zapcore.Level) string {
	ctx, buf := test.TestableContext(logLevel)
	ctx, s := spiffylogger.OpenCustomSpan(ctx, "rootSpan")

	s.Debug("testDebug")
	s.Debugf("%s", "testDebugf")
	s.Info("testInfo")
	s.Infof("%s", "testInfof")
	s.Error(errors.New("testError"))
	s.Errorf(errors.New("testErrorf"), "%s", "testGoErr")

	_, cs := spiffylogger.OpenSpan(ctx)
	cs.Info("hello!")

	s.Close()
	return buf.String()
}

func verifyAllJSON(t *testing.T, logged string) {
	lines := strings.Split(logged, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// test each line is valid json
		ll := &spiffylogger.LogLine{}
		err := json.Unmarshal([]byte(line), ll)
		require.NoError(t, err, "failed to unmarshal json log line %s", line)
	}
}

func TestHumanOutput(t *testing.T) {
	ctx, buf := test.TestableContext(zapcore.ErrorLevel)
	_, s := spiffylogger.OpenSpan(ctx)

	err := externalFunc()
	s.Error(err)

	s.Close()

	result := buf.String()
	require.Contains(t, result, "msg=root cause\n")
	require.Contains(t, result, "external context\n")
	require.Contains(t, result, "code.chipmunq.com/internal/main_test.externalFunc")
	require.Contains(t, result, "/internal/span_test.go:")
	require.Contains(t, result, "code.chipmunq.com/internal/main_test.TestHumanOutput")
	require.Contains(t, result, "testing.tRunner\n")
	require.Contains(t, result, "runtime.goexit\n")
}

func externalFunc() error {
	return errors.Wrap(func() error { return fmt.Errorf("root cause") }(), "external context")
}
