package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func Test_Config(t *testing.T) {
	c := NewConfig("appName")
	cmd := cobra.Command{}
	c.AddFlags(cmd.Flags())

	require.NotNil(t, c.Logger)
	require.Equal(t, "appName", c.AppName)

	var flags []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, flag.Name)
	})

	require.Contains(t, flags, "config")
}
