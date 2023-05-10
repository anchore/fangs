package fangs

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
)

type FlagAdder interface {
	AddFlags(flags *pflag.FlagSet)
}

func AddFlags(flags *pflag.FlagSet, structs ...any) {
	f := reflect.ValueOf(flags)
	for _, o := range structs {
		v := reflect.ValueOf(o)
		if !isPtr(v.Type()) {
			panic(fmt.Sprintf("AddFlags must be called with pointer receviers, got: %#v", o))
		}
		addFlags(f, v)
	}
}

func addFlags(flags reflect.Value, v reflect.Value) {
	invokeAddFlags(flags, v)

	v, t := base(v)

	if isStruct(t) {
		for i := 0; i < t.NumField(); i++ {
			v := v.Field(i)
			v = v.Addr()
			if !v.CanInterface() {
				continue
			}

			addFlags(flags, v)
		}
	}
}

func invokeAddFlags(flags reflect.Value, v reflect.Value) {
	defer func() {
		// we need to handle embedded structs having AddFlags methods called, adding flags with existing names
		// FIXME: bad idea, should at least log something
		_ = recover()
	}()

	t := v.Type()
	m, ok := t.MethodByName("AddFlags")

	if ok {
		_ = m.Func.Call([]reflect.Value{v, flags})
	}
}
