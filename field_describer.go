package fangs

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
)

// FieldDescriber a struct implementing this interface will have DescribeFields called when Summarize is called
type FieldDescriber interface {
	DescribeFields(descriptions FieldDescriptionSet)
}

// FieldDescriptionSet accepts field descriptions
type FieldDescriptionSet interface {
	Add(ptr any, description string)
}

// FieldDescriptionSetProvider implements both DescriptionProvider and FieldDescriptionSet
type FieldDescriptionSetProvider interface {
	DescriptionProvider
	FieldDescriptionSet
}

type directDescriber struct {
	flagRefs flagRefs
}

var _ FieldDescriptionSetProvider = (*directDescriber)(nil)

func NewDirectDescriber() FieldDescriptionSetProvider {
	return &directDescriber{
		flagRefs: flagRefs{},
	}
}

func NewFieldDescriber(cfgs ...any) DescriptionProvider {
	d := NewDirectDescriber()
	for _, v := range cfgs {
		addFieldDescriptions(d, v)
	}
	return d
}

func (d *directDescriber) Add(ptr any, description string) {
	v := reflect.ValueOf(ptr)
	if !isPtr(v.Type()) {
		panic(fmt.Sprintf("Add() requires a pointer, but got: %#v", ptr))
	}
	p := v.Pointer()
	d.flagRefs[p] = &pflag.Flag{
		Usage: description,
	}
}

func (d *directDescriber) GetDescription(v reflect.Value, _ reflect.StructField) string {
	if v.CanAddr() {
		v = v.Addr()
	}
	if isPtr(v.Type()) {
		f := d.flagRefs[v.Pointer()]
		if f != nil {
			return f.Usage
		}
	}
	return ""
}

func addFieldDescriptions(d FieldDescriptionSet, o any) {
	v := reflect.ValueOf(o)
	t := v.Type()
	if isPtr(t) {
		if p, ok := o.(FieldDescriber); ok && !isPromotedMethod(o, "DescribeFields") {
			p.DescribeFields(d)
		}
		v, t = base(v)
	}

	if !isStruct(t) {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		v := v.Field(i)
		t := v.Type()
		if isPtr(t) {
			v = v.Elem()
			t = v.Type()
		}
		if !v.CanAddr() || !isStruct(t) {
			continue
		}
		addFieldDescriptions(d, v.Addr().Interface())
	}
}
