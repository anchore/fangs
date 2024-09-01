package fangs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_Ptr(t *testing.T) {
	cmd := &cobra.Command{}

	type typ struct {
		BoolVal    *bool
		NotBoolVal *bool
		IntVal     *int
		StringVal  *string
		FloatVal   *float64
	}

	a := &typ{
		NotBoolVal: p(true),
	}

	flags := cmd.Flags()
	BoolPtrVarP(flags, &a.BoolVal, "bool-ptr", "", "bool ptr usage")
	BoolPtrVarP(flags, &a.NotBoolVal, "not-bool-ptr", "", "not bool ptr usage")
	IntPtrVarP(flags, &a.IntVal, "int-ptr", "", "int ptr usage")
	StringPtrVarP(flags, &a.StringVal, "string-ptr", "", "string ptr usage")
	Float64PtrVarP(flags, &a.FloatVal, "float-ptr", "", "float ptr usage")

	require.Nil(t, a.BoolVal)
	require.Nil(t, a.IntVal)
	require.Nil(t, a.StringVal)

	err := flags.Parse([]string{
		"--bool-ptr",
		"--not-bool-ptr",
		"--int-ptr", "17",
		"--string-ptr", "some-string",
		"--float-ptr", "64.8",
	})
	require.NoError(t, err)

	require.NotNil(t, a.BoolVal)
	require.Equal(t, true, *a.BoolVal)

	require.NotNil(t, a.NotBoolVal)
	require.Equal(t, false, *a.NotBoolVal)

	require.NotNil(t, a.IntVal)
	require.Equal(t, 17, *a.IntVal)

	require.NotNil(t, a.StringVal)
	require.Equal(t, "some-string", *a.StringVal)

	require.NotNil(t, a.FloatVal)
	require.Equal(t, 64.8, *a.FloatVal)
}
