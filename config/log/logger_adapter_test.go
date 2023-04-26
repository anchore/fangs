package log

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/anchore/go-logger"
)

func Test_loggerAdapter(t *testing.T) {
	buf := &bytes.Buffer{}
	i := impl{
		writer: buf,
	}
	l := FromLogger(i)

	l.Error("error ")
	l.Warn("warn ")
	l.Debug("debug ")
	l.Trace("trace ")

	require.Equal(t, "error warn debug trace ", buf.String())
}

type impl struct {
	writer io.Writer
}

func (i impl) Errorf(format string, args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprintf(format, args...)))
}

func (i impl) Error(args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprint(args...)))
}

func (i impl) Warnf(format string, args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprintf(format, args...)))
}

func (i impl) Warn(args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprint(args...)))
}

func (i impl) Infof(format string, args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprintf(format, args...)))
}

func (i impl) Info(args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprint(args...)))
}

func (i impl) Debugf(format string, args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprintf(format, args...)))
}

func (i impl) Debug(args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprint(args...)))
}

func (i impl) Tracef(format string, args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprintf(format, args...)))
}

func (i impl) Trace(args ...interface{}) {
	_, _ = i.writer.Write([]byte(fmt.Sprint(args...)))
}

func (i impl) WithFields(_ ...interface{}) logger.MessageLogger {
	return i
}

func (i impl) Nested(_ ...interface{}) logger.Logger {
	return i
}

var _ logger.Logger = (*impl)(nil)
