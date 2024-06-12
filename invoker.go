package fangs

import (
	"fmt"
	"reflect"
)

// InvokeAll recursively calls the invoker function with anything implementing the interface in the object graph.
// the type of the parameter to the invoker function is used to determine the interface, which must have exactly one
// method. InvokeAll will also avoid duplicate calls to methods on embedded structs where the method is inherited.
// InvokeAll optionally creates empty structs at every location in the object graph where a nil value exists that would
// point to a struct type; this may be used to ensure certain calls such as AddFlags and Summarize will always reference
// the same objects in memory.
func InvokeAll[T any](obj any, invoker func(T) error, opts ...func(*invokeAll)) error {
	invokerFunc := reflect.ValueOf(invoker)
	// get the target interface type
	invokerFuncType := invokerFunc.Type()
	interfaceType := invokerFuncType.In(0) // must have exactly 1 argument per func signature
	iv := invokeAll{
		interfaceType: interfaceType,
		invokeFunc:    invokerFunc,
		funcName:      funcName(interfaceType),
	}
	for _, opt := range opts {
		opt(&iv)
	}
	return iv.invokeAll(reflect.ValueOf(obj))
}

// InvokeAllCreateStructs is an option to InvokeAll which causes nil structs pointers to be automatically populated with
// empty values
func InvokeAllCreateStructs(iv *invokeAll) {
	iv.createStructs = true
}

// InvokeAllRequirePtr is an option to InvokeAll that indicates interface implementations must have a pointer receiver
func InvokeAllRequirePtr(iv *invokeAll) {
	iv.requirePtr = true
}

func funcName(interfaceType reflect.Type) string {
	if interfaceType.NumMethod() != 1 {
		panic(fmt.Sprintf("provided interfaces must have exactly 1 method, got %v", interfaceType.NumMethod()))
	}
	m := interfaceType.Method(0)
	return m.Name
}

type invokeAll struct {
	interfaceType reflect.Type
	invokeFunc    reflect.Value
	funcName      string
	createStructs bool
	requirePtr    bool
}

func (iv *invokeAll) invoke(v reflect.Value) error {
	out := iv.invokeFunc.Call([]reflect.Value{v})[0] // must have exactly 1 error return value per func signature
	if out.IsNil() {
		return nil
	}
	return out.Interface().(error)
}

func (iv *invokeAll) invokeAll(v reflect.Value) error {
	t := v.Type()

	for isPtr(t) {
		if v.IsNil() {
			return nil
		}

		if v.CanInterface() {
			if v.Type().Implements(iv.interfaceType) && !isPromotedMethod(v, iv.funcName) {
				if err := iv.invoke(v); err != nil {
					return err
				}
			}
		}
		t = t.Elem()
		v = v.Elem()
	}

	// fail if implements the interface with something not using a pointer receiver
	if v.Type().Implements(iv.interfaceType) && !isPromotedMethod(v, iv.funcName) {
		if iv.requirePtr {
			return fmt.Errorf("type implements interface without pointer reference: %v implements %v", v.Type(), iv.interfaceType)
		}
		if err := iv.invoke(v); err != nil {
			return err
		}
	}

	switch {
	case isStruct(t):
		return iv.invokeAllStruct(v)
	case isSlice(t):
		return iv.invokeAllSlice(v)
	case isMap(t):
		return iv.invokeAllMap(v)
	}

	return nil
}

// invokeAllStruct call recursively on struct fields
func (iv *invokeAll) invokeAllStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if !includeField(f) {
			continue
		}

		v := v.Field(i)

		if isNil(v) {
			// optionally create structs when there is only a nil pointer to it
			if iv.createStructs && isStruct(v.Type().Elem()) {
				fv := reflect.New(v.Type().Elem())
				v.Set(fv) // set the newly created struct
				v = fv
			} else {
				continue
			}
		}

		for isPtr(v.Type()) {
			v = v.Elem()
		}

		if !v.CanAddr() {
			continue
		}

		if err := iv.invokeAll(v.Addr()); err != nil {
			return err
		}
	}
	return nil
}

// invokeAllSlice call recursively on slice items
func (iv *invokeAll) invokeAllSlice(v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		v := v.Index(i)

		if isNil(v) {
			continue
		}

		for isPtr(v.Type()) {
			v = v.Elem()
		}

		if !v.CanAddr() {
			continue
		}

		if err := iv.invokeAll(v.Addr()); err != nil {
			return err
		}
	}
	return nil
}

// invokeAllMap call recursively on map values
func (iv *invokeAll) invokeAllMap(v reflect.Value) error {
	mapV := v
	i := v.MapRange()
	for i.Next() {
		v := i.Value()

		if isNil(v) {
			continue
		}

		for isPtr(v.Type()) {
			v = v.Elem()
		}

		if !v.CanAddr() {
			// unable to call .Addr() on struct map entries, so copy to a new instance and set on the map
			if isStruct(v.Type()) {
				newV := reflect.New(v.Type())
				newV.Elem().Set(v)
				if err := iv.invokeAll(newV); err != nil {
					return err
				}
				mapV.SetMapIndex(i.Key(), newV.Elem())
			}

			continue
		}

		if err := iv.invokeAll(v.Addr()); err != nil {
			return err
		}
	}
	return nil
}
