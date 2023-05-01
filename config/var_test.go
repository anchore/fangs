package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_Ptr(t *testing.T) {
	cmd := &cobra.Command{}

	type typ struct {
		BoolVal   *bool
		IntVal    *int
		StringVal *string
	}

	a := &typ{}

	flags := cmd.Flags()
	BoolPtrVarP(flags, &a.BoolVal, "bool-ptr", "", "bool ptr usage")
	IntPtrVarP(flags, &a.IntVal, "int-ptr", "", "int ptr usage")
	StringPtrVarP(flags, &a.StringVal, "string-ptr", "", "string ptr usage")

	require.Nil(t, a.BoolVal)
	require.Nil(t, a.IntVal)
	require.Nil(t, a.StringVal)

	err := flags.Set("bool-ptr", "true")
	require.NoError(t, err)
	require.NotNil(t, a.BoolVal)
	require.Equal(t, true, *a.BoolVal)

	err = flags.Set("int-ptr", "17")
	require.NoError(t, err)
	require.NotNil(t, a.IntVal)
	require.Equal(t, 17, *a.IntVal)

	err = flags.Set("string-ptr", "some-string")
	require.NoError(t, err)
	require.NotNil(t, a.StringVal)
	require.Equal(t, "some-string", *a.StringVal)
}
