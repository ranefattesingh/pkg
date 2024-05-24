package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	zap *zap.Logger
}

type Encoder zapcore.Encoder

type Config struct {
	Output           io.Writer
	LogLevel         LogLevel
	Encoder          Encoder
	AdditionalFields map[string]any
	IsDevelopment    bool
}

var log *logger

func Init(c Config) {
	if log == nil {
		log = &logger{
			zap: newZapLogger(c),
		}
	}
}

func newZapLogger(c Config) *zap.Logger {
	output := c.Output
	if output == nil {
		output = os.Stdout
	}

	if c.Encoder == nil {
		c.Encoder = DefaultJSONEncoder()
	}

	logLevel, err := zapcore.ParseLevel(string(c.LogLevel))
	if err != nil {
		panic("invalid log level")
	}

	options := []zap.Option{
		zap.WithCaller(true),
		zap.AddStacktrace(stacktraceEnabler{}),
	}

	if c.IsDevelopment {
		options = append(options, zap.Development())
	}

	core := zapcore.NewCore(c.Encoder, zapcore.AddSync(output), zap.NewAtomicLevelAt(logLevel))
	zapLogger := zap.New(core, options...)

	return zapLogger
}

type stacktraceEnabler struct{}

func (s stacktraceEnabler) Enabled(level zapcore.Level) bool {
	switch level {
	case zap.PanicLevel, zap.DPanicLevel, zapcore.ErrorLevel:
		return true
	default:
		return false
	}
}

func DefaultJSONEncoder() Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		FunctionKey:    zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func DefaultConsoleEncoder() Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "L",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		FunctionKey:    zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func Logger() *logger {
	return log
}
