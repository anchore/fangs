package fangs

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/anchore/go-logger/adapter/discard"
)

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

	assert.Equal(t, flagNames, []string{"sub2-flag"})
}

func Test_AddFlags(t *testing.T) {
	flags := pflag.NewFlagSet("set", pflag.ContinueOnError)
	t1 := &T1{}
	AddFlags(discard.New(), flags, t1)

	var flagNames []string
	flags.VisitAll(func(flag *pflag.Flag) {
		flagNames = append(flagNames, flag.Name)
	})

	assert.Len(t, flagNames, 3)
	assert.Contains(t, flagNames, "t1-flag")
	assert.Contains(t, flagNames, "sub2-flag")
	assert.Contains(t, flagNames, "sub3-flag")
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
