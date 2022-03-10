package test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/bhechinger/spiffylogger"
)

/*
The goal of this file is to create a variety of test helpers.
When you find yourself needing extra log-parsing functionality, add it here!
*/

// JSONLogToLines turns a buffered log (JSON output only) into a series of log lines
func JSONLogToLines(t *testing.T, b *bytes.Buffer) []*spiffylogger.LogLine {
	lls := make([]*spiffylogger.LogLine, 0)

	ss := strings.Split(b.String(), "\n")
	for _, sl := range ss {
		ll := &spiffylogger.LogLine{}
		err := json.Unmarshal([]byte(sl), ll)
		require.NoError(t, err, "failed to unmarshal log line")
		lls = append(lls, ll)
	}

	return lls
}

// LogContainsLine checks for a line containing a specific log level, log name, and message containing
func LogContainsLine(t *testing.T, lls []*spiffylogger.LogLine, level zapcore.Level, name string, msg string) {
	found := false
	for _, ll := range lls {
		if ll.Level == level && ll.Name == name {
			if strings.Contains(ll.Message, msg) {
				found = true
				break
			}
		}
	}
	require.True(t, found, "failed to find log line with level=%d and name=%s, and msg containing %s", level, name, msg)
}

// LogContainsLineWithMsg checks for a log line containing the given message
func LogContainsLineWithMsg(t *testing.T, lls []*spiffylogger.LogLine, msg string) {
	found := false
	for _, ll := range lls {
		if strings.Contains(ll.Message, msg) {
			found = true
			break
		}
	}
	require.True(t, found, "failed to find log line with msg containing %s", msg)
}
