package log

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel LogLevel = "debug"
	// InfoLevel is the default logging priority.
	InfoLevel LogLevel = "info"
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel LogLevel = "warn"
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel LogLevel = "error"
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel LogLevel = "dpanic"
	// PanicLevel logs a message, then panics.
	PanicLevel LogLevel = "panic"
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel LogLevel = "fatal"
)

type LogLevel string
