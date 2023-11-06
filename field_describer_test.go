package fangs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fieldDescriber(t *testing.T) {
	f1 := &fdTest1{
		Ptr: &fdTest3{},
	}

	d := NewFieldDescriber(f1)

	require.Equal(t, 1, f1.called)
	require.Equal(t, 1, f1.FdTest2.called)
	require.Equal(t, 1, f1.Ptr.called)

	dd := d.(*directDescriber)

	var values []string
	for _, d := range dd.flagRefs {
		values = append(values, d.Usage)
	}

	require.Contains(t, values, "string description")
	require.Contains(t, values, "int description")
	require.Contains(t, values, "fd test 3 value description")
}

func Test_FieldDescriberDoesNotPanicOnEmbeddedUnexportedStruct(t *testing.T) {
	type moduleConfig struct {
		ModuleBool bool `yaml:"module-bool" mapstructure:"module-bool"`
	}

	type specialModuleConfig struct {
		moduleConfig      `yaml:",inline" mapstructure:",squash"`
		SpecialModuleBool bool `yaml:"special-module-bool" mapstructure:"special-module-bool"`
	}

	type TopLevelConfig struct {
		Module1 moduleConfig        `yaml:"module-1" mapstructure:"module-1"`
		Module2 specialModuleConfig `yaml:"module-2" mapstructure:"module-2"`
	}

	cfgPtr := &TopLevelConfig{}
	_ = NewFieldDescriber(cfgPtr)
}

func Test_FieldDescriberDoesNotPanicOnEmbeddedExportedStructPointer(t *testing.T) {
	type ModuleConfig struct {
		ModuleBool bool `yaml:"module-bool" mapstructure:"module-bool"`
	}

	type specialModuleConfig struct {
		*ModuleConfig     `yaml:",inline" mapstructure:",squash"`
		SpecialModuleBool bool `yaml:"special-module-bool" mapstructure:"special-module-bool"`
	}

	type TopLevelConfig struct {
		Module1 ModuleConfig        `yaml:"module-1" mapstructure:"module-1"`
		Module2 specialModuleConfig `yaml:"module-2" mapstructure:"module-2"`
	}

	cfgPtr := &TopLevelConfig{}
	_ = NewFieldDescriber(cfgPtr)
}

func Test_FieldDescriberDoesNotPanicOnEmbeddedUnexportedStructPointer(t *testing.T) {
	type moduleConfig struct {
		ModuleBool bool `yaml:"module-bool" mapstructure:"module-bool"`
	}

	type specialModuleConfig struct {
		*moduleConfig     `yaml:",inline" mapstructure:",squash"`
		SpecialModuleBool bool `yaml:"special-module-bool" mapstructure:"special-module-bool"`
	}

	type TopLevelConfig struct {
		Module1 moduleConfig        `yaml:"module-1" mapstructure:"module-1"`
		Module2 specialModuleConfig `yaml:"module-2" mapstructure:"module-2"`
	}

	cfgPtr := &TopLevelConfig{}
	_ = NewFieldDescriber(cfgPtr)
}

type fdTest1 struct {
	called  int
	String  string
	FdTest2 fdTest2
	Ptr     *fdTest3
}

func (f *fdTest1) DescribeFields(d FieldDescriptionSet) {
	f.called++
	d.Add(&f.String, "string description")
}

var _ FieldDescriber = &fdTest1{}

type fdTest2 struct {
	called int
	Int    int
}

func (f *fdTest2) DescribeFields(d FieldDescriptionSet) {
	f.called++
	d.Add(&f.Int, "int description")
}

var _ FieldDescriber = &fdTest2{}

type fdTest3 struct {
	called int
	Value  string
}

func (f *fdTest3) DescribeFields(d FieldDescriptionSet) {
	f.called++
	d.Add(&f.Value, "fd test 3 value description")
}

var _ FieldDescriber = &fdTest3{}
