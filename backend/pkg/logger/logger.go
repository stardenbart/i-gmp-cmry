package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger for structured logging.
type Logger struct {
	*zap.Logger
}

// New creates a new Logger based on config values.
func New(level, output, filePath string) *Logger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var core zapcore.Core

	if output == "file" {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("logger: cannot open log file: " + err.Error())
		}
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
		core = zapcore.NewCore(fileEncoder, zapcore.AddSync(f), zapLevel)
	} else {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
		core = zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel)
	}

	return &Logger{zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))}
}

// ─── Convenience field constructors ───────────────────────────────────────

func String(key, val string) zap.Field  { return zap.String(key, val) }
func Int(key string, val int) zap.Field { return zap.Int(key, val) }
func Error(err error) zap.Field         { return zap.Error(err) }
func Any(key string, val interface{}) zap.Field { return zap.Any(key, val) }
