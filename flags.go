package fangs

import (
	"github.com/spf13/pflag"

	"github.com/anchore/go-logger"
)

// FlagAdder interface can be implemented by structs in order to add flags when AddFlags is called
type FlagAdder interface {
	AddFlags(flags FlagSet)
}

// AddFlags traverses the object graphs from the structs provided and calls all AddFlags methods implemented on them
func AddFlags(log logger.Logger, flags *pflag.FlagSet, structs ...any) {
	flagSet := NewPFlagSet(log, flags)
	for _, o := range structs {
		_ = InvokeAll(o, func(flagAdder FlagAdder) error {
			flagAdder.AddFlags(flagSet)
			return nil
		}, InvokeAllCreateStructs, InvokeAllRequirePtr)
	}
}
