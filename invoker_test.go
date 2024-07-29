package fangs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_invoker(t *testing.T) {
	calls := 0
	s := badStructImpl{
		incr: func() {
			calls++
		},
	}

	invoker := func(l PostLoader) error {
		return l.PostLoad()
	}

	err := InvokeAll(s, invoker)
	require.NoError(t, err)
	require.Equal(t, 1, calls)

	err = InvokeAll(s, invoker, InvokeAllRequirePtr)
	require.Error(t, err)
	require.Equal(t, 1, calls)
}

type badStructImpl struct {
	incr func()
}

func (s badStructImpl) PostLoad() error {
	s.incr()
	return nil
}
