package log

import "go.uber.org/zap/zapcore"

func Info(message string, fields ...zapcore.Field) {
	log.zap.Info(message, fields...)
}

func Error(message string, fields ...zapcore.Field) {
	log.zap.Error(message, fields...)
}

func Fatal(message string, fields ...zapcore.Field) {
	log.zap.Fatal(message, fields...)
}

func Panic(message string, fields ...zapcore.Field) {
	log.zap.Panic(message, fields...)
}

func DPanic(message string, fields ...zapcore.Field) {
	log.zap.DPanic(message, fields...)
}

func Debug(message string, fields ...zapcore.Field) {
	log.zap.Debug(message, fields...)
}
