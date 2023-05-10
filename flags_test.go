package fangs

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func Test_AddFlags(t *testing.T) {
	flags := pflag.NewFlagSet("set", pflag.ContinueOnError)
	t1 := &T1{}
	AddFlags(flags, t1)

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

func (t *Sub2) AddFlags(flags *pflag.FlagSet) {
	flags.String("sub2-flag", "val", "usage")
}

type Sub3 struct {
	Value string
}

func (t Sub3) AddFlags(flags *pflag.FlagSet) {
	flags.String("sub3-flag", "val", "usage")
}

type Sub1 struct {
	Sub2
	S3 Sub3
}

type T1 struct {
	Val Sub1
}

func (t T1) AddFlags(flags *pflag.FlagSet) {
	flags.Int("t1-flag", 1, "usage")
}

var _ FlagAdder = (*T1)(nil)
