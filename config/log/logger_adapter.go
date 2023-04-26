package log

import (
	"github.com/anchore/go-logger"
)

// FromLogger returns a log adapter from the anchore/go-logger
func FromLogger(l logger.Logger) Log {
	return loggerAdapter{
		l: l,
	}
}

type loggerAdapter struct {
	l logger.Logger
}

func (l loggerAdapter) Error(message string) {
	l.l.Error(message)
}

func (l loggerAdapter) Warn(message string) {
	l.l.Warn(message)
}

func (l loggerAdapter) Debug(message string) {
	l.l.Debug(message)
}

func (l loggerAdapter) Trace(message string) {
	l.l.Trace(message)
}

var _ Log = (*loggerAdapter)(nil)
