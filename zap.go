package spiffylogger

import (
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level zapcore.Level, options ...zap.Option) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(level)
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(t.Format(time.RFC3339Nano))
	}

	opts := []zap.Option{
		zap.AddCallerSkip(2),
	}
	opts = append(opts, options...)

	zapLogger, err := cfg.Build(opts...)
	if err != nil {
		log.Fatalf("err creating logger: %v\n", err.Error())
	}

	return zapLogger
}

// // Error implements the LeveledLogWriter Error func with a zap logger.
// func (zl LeveledLogger) Error(ll LogLine) {
// 	zl.Logger.Error(ll.Message, ll.ZapFields()...)
// }
//
// // Warn implements the LeveledLogWriter Warn func with a zap logger.
// func (zl LeveledLogger) Warn(ll LogLine) {
// 	zl.Logger.Warn(ll.Message, ll.ZapFields()...)
// }
//
// // Info implements the LeveledLogWriter Info func with a zap logger.
// func (zl LeveledLogger) Info(ll LogLine) {
// 	zl.Logger.Info(ll.Message, ll.ZapFields()...)
// }
//
// // Debug implements the LeveledLogWriter Debug func with a zap logger.
// func (zl LeveledLogger) Debug(ll LogLine) {
// 	zl.Logger.Debug(ll.Message, ll.ZapFields()...)
// }

// ZapFields converts a LogLine to a slice of zapcore.Field.
//
// Zap already has built in fields for these log line information:
// - ll.Timestamp	=> ts
// - ll.File		=> caller
// - ll.LineNumber	=> caller
func (ll LogLine) ZapFields(duration int64) []zapcore.Field {
	zapFields := []zapcore.Field{
		zap.String("name", ll.Name),
		zap.String("correlation_id", ll.CorrelationID),
		zap.String("span_id", ll.SpanID),
		zap.Int64("duration", duration),
	}

	return append(zapFields, ll.Fields...)
}
