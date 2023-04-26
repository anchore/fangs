package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_discardTest(t *testing.T) {
	require.NotPanics(t, func() {
		l := NewDiscard()

		l.Error("error ")
		l.Warn("warn ")
		l.Debug("debug ")
		l.Trace("trace ")
	})
}
