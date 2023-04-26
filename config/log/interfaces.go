package log

type Log interface {
	Error(message string)
	Warn(message string)
	Debug(message string)
	Trace(message string)
}
