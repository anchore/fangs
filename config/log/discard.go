package log

func NewDiscard() Log {
	return discard{}
}

type discard struct{}

func (d discard) Error(_ string) {}

func (d discard) Warn(_ string) {}

func (d discard) Debug(_ string) {}

func (d discard) Trace(_ string) {}

var _ Log = (*discard)(nil)
