package log

func Info(message string, fields ...Field) {
	log.zap.Info(message, toZapFields(fields...)...)
}

func Error(message string, fields ...Field) {
	log.zap.Error(message, toZapFields(fields...)...)
}

func Fatal(message string, fields ...Field) {
	log.zap.Fatal(message, toZapFields(fields...)...)
}

func Panic(message string, fields ...Field) {
	log.zap.Panic(message, toZapFields(fields...)...)
}

func DPanic(message string, fields ...Field) {
	log.zap.DPanic(message, toZapFields(fields...)...)
}

func Debug(message string, fields ...Field) {
	log.zap.Debug(message, toZapFields(fields...)...)
}
