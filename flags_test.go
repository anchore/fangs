package fangs

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anchore/go-logger/adapter/discard"
)

func Test_PFlagSetProvider(t *testing.T) {
	flags := pflag.NewFlagSet("set", pflag.ContinueOnError)
	flagSet := NewPFlagSet(discard.New(), flags)
	prov, ok := flagSet.(PFlagSetProvider)
	require.True(t, ok)
	require.Equal(t, flags, prov.PFlagSet())
}

func Test_EmbeddedAddFlags(t *testing.T) {
	type ty1 struct {
		Something string
		Sub2
	}

	flags := pflag.NewFlagSet("set", pflag.ContinueOnError)
	t1 := &ty1{}

	AddFlags(discard.New(), flags, t1)

	var flagNames []string
	flags.VisitAll(func(flag *pflag.Flag) {
		flagNames = append(flagNames, flag.Name)
	})

	require.Equal(t, flagNames, []string{"sub2-flag"})
}

func Test_AddFlags(t *testing.T) {
	flags := pflag.NewFlagSet("set", pflag.ContinueOnError)
	t1 := &T1{}
	AddFlags(discard.New(), flags, t1)

	var flagNames []string
	flags.VisitAll(func(flag *pflag.Flag) {
		flagNames = append(flagNames, flag.Name)
	})

	require.Len(t, flagNames, 3)
	require.Contains(t, flagNames, "t1-flag")
	require.Contains(t, flagNames, "sub2-flag")
	require.Contains(t, flagNames, "sub3-flag")
}

func Test_AddFlags_StructRefs(t *testing.T) {
	flags := pflag.NewFlagSet("set", pflag.ContinueOnError)

	type ty2 struct {
		Nested   string
		Optional *bool // ensure this is not set
	}
	type ty1 struct {
		T2 *ty2 // ensure the zero value is set
	}

	t1 := &ty1{}

	AddFlags(discard.New(), flags, t1)

	require.NotNil(t, t1.T2)
	assert.Nil(t, t1.T2.Optional)
}

type Sub2 struct {
	Value string
}

func (t *Sub2) AddFlags(flags FlagSet) {
	flags.StringVarP(&t.Value, "sub2-flag", "", "usage")
}

type Sub3 struct {
	Value string
}

func (t *Sub3) AddFlags(flags FlagSet) {
	flags.StringVarP(&t.Value, "sub3-flag", "", "usage")
}

var _ FlagAdder = (*Sub3)(nil)

type Sub1 struct {
	Sub2
	S3 Sub3
}

type T1 struct {
	Ival int
	Val  Sub1
}

func (t *T1) AddFlags(flags FlagSet) {
	flags.IntVarP(&t.Ival, "t1-flag", "", "usage")
}

var _ FlagAdder = (*T1)(nil)
