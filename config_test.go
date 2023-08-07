package fangs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/anchore/go-logger/adapter/discard"
)

func Test_BasicConfig(t *testing.T) {
	c := NewConfig("appName")
	cmd := cobra.Command{}

	fs := NewPFlagSet(discard.New(), cmd.Flags())
	c.AddFlags(fs)

	require.NotNil(t, c.Logger)
	require.Equal(t, "appName", c.AppName)

	var flags []string
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flags = append(flags, flag.Name)
	})

	require.Contains(t, flags, "config")
}

func Test_EnvVarConfig(t *testing.T) {
	t.Setenv("APPNAME_CONFIG", "some/config.env")

	c := NewConfig("appName").WithConfigEnvVar()
	require.Equal(t, c.File, "some/config.env")

	cmd := cobra.Command{}

	fs := NewPFlagSet(discard.New(), cmd.Flags())
	c.AddFlags(fs)

	// simulate the flag set
	err := cmd.Flags().Set("config", "a/config.flag")
	require.NoError(t, err)
	require.Equal(t, c.File, "a/config.flag")
}
