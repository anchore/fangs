package fangs

import (
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AddsFlags interface {
	AddFlags(flags *pflag.FlagSet)
}

func AddCommandFlags(cmd *cobra.Command, structs ...any) {
	AddFlags(cmd.Flags(), structs...)
}

func AddFlags(flags *pflag.FlagSet, structs ...any) {
	for _, o := range structs {
		addFlags(flags, false, o)
	}
}

func addFlags(flags *pflag.FlagSet, skip bool, o any) {
	if !skip {
		invokeAddFlags(flags, o)
	}

	v, t := base(reflect.ValueOf(o))

	if isStruct(t) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			v := v.Field(i)
			v = v.Addr()
			if !v.CanInterface() {
				continue
			}

			addFlags(flags, f.Anonymous, v.Interface())
		}
	}
}

func invokeAddFlags(flags *pflag.FlagSet, o any) {
	if o, ok := o.(AddsFlags); ok {
		o.AddFlags(flags)
		return
	}
	v := reflect.ValueOf(o)
	if isPtr(v.Type()) {
		v = v.Elem()
		if v.CanInterface() {
			invokeAddFlags(flags, v.Interface())
		}
	}
}
