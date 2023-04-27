package config

import (
	"os"
	"path"
	"testing"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sub struct {
	Sv      string `json:"sv" yaml:"sv" mapstructure:"sv"`
	Unbound string `json:"unbound" yaml:"unbound" mapstructure:"unbound"`
}

type root struct {
	V   string `json:"v" yaml:"v" mapstructure:"v"`
	Sub *sub   `json:"sub" yaml:"sub" mapstructure:"sub"`
}

func Test_LoadDefaults(t *testing.T) {
	cmd, cfg, r, s := setup(t)

	err := Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "default-sv", s.Sv)
	require.Equal(t, "default-v", r.V)
}

func Test_LoadFromConfigFile(t *testing.T) {
	cmd, cfg, r, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.ConfigFile = path.Join(wd, "test-fixtures", "config.yaml")

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "direct-config-sub-v", s.Sv)
	require.Equal(t, "direct-config-v", r.V)
}

func Test_LoadFromEnv(t *testing.T) {
	t.Setenv("MY_APP_V", "env-var-v")
	t.Setenv("MY_APP_SUB_SV", "env-var-sv")

	cmd, cfg, r, s := setup(t)

	err := Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "env-var-sv", s.Sv)
	require.Equal(t, "env-var-v", r.V)
}

func Test_LoadFromEnvOverridingConfigFile(t *testing.T) {
	t.Setenv("MY_APP_V", "env-var-v")
	t.Setenv("MY_APP_SUB_SV", "env-var-sv")

	cmd, cfg, r, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.ConfigFile = path.Join(wd, "test-fixtures", "config.yaml")

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "env-var-sv", s.Sv)
	require.Equal(t, "env-var-v", r.V)
}

func Test_LoadSubStruct(t *testing.T) {
	t.Setenv("MY_APP_SUB_SV", "env-var-sv")

	cmd, cfg, _, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.ConfigFile = path.Join(wd, "test-fixtures", "config.yaml")

	err = LoadAt(cfg, cmd, "sub", s)
	require.NoError(t, err)

	require.Equal(t, "env-var-sv", s.Sv)
}

func Test_LoadSubStructEnv(t *testing.T) {
	cmd, cfg, _, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.ConfigFile = path.Join(wd, "test-fixtures", "config.yaml")

	err = LoadAt(cfg, cmd, "sub", s)
	require.NoError(t, err)

	require.Equal(t, "direct-config-sub-v", s.Sv)
}

func Test_LoadFromFlags(t *testing.T) {
	cmd, cfg, r, s := setup(t)

	err := cmd.PersistentFlags().Set("v", "flag-value-v")
	require.NoError(t, err)

	err = cmd.Flags().Set("sv", "flag-value-sv")
	require.NoError(t, err)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "flag-value-sv", s.Sv)
	require.Equal(t, "flag-value-v", r.V)
}

func Test_LoadFromFlagsOverridingAll(t *testing.T) {
	t.Setenv("MY_APP_V", "env-var-v")
	t.Setenv("MY_APP_SUB_SV", "env-var-sv")

	cmd, cfg, r, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.ConfigFile = path.Join(wd, "test-fixtures", "config.yaml")

	err = cmd.PersistentFlags().Set("v", "flag-value-v")
	require.NoError(t, err)

	err = cmd.Flags().Set("sv", "flag-value-sv")
	require.NoError(t, err)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "flag-value-sv", s.Sv)
	require.Equal(t, "flag-value-v", r.V)
}

func setup(_ *testing.T) (*cobra.Command, Config, *root, *sub) {
	cfg := NewConfig("my-app")

	s := &sub{
		Sv:      "default-sv",
		Unbound: "default-unbound",
	}

	r := &root{
		V:   "default-v",
		Sub: s,
	}

	cmd := &cobra.Command{}

	flags := cmd.PersistentFlags()
	flags.StringVarP(&r.V, "v", "", r.V, "v usage")

	flags = cmd.Flags()
	flags.StringVarP(&s.Sv, "sv", "", s.Sv, "sv usage")

	return cmd, cfg, r, s
}

func Test_AllFieldTypes(t *testing.T) {
	t.Setenv("APP_BOOL", "true")
	t.Setenv("APP_STRING", "stringValueEnv")
	t.Setenv("APP_STRING_ARRAY", "stringArrayValueEnv")

	type all struct {
		Bool        bool     `mapstructure:"bool"`
		String      string   `mapstructure:"string"`
		StringArray []string `mapstructure:"string-array"`
	}

	a := &all{
		String:      "stringValue",
		StringArray: []string{"stringArrayValue"},
	}

	cfg := NewConfig("app")

	cmd := &cobra.Command{}

	flags := cmd.Flags()
	flags.BoolVarP(&a.Bool, "bool", "", a.Bool, "bool usage")
	flags.StringVarP(&a.String, "string", "", a.String, "string usage")
	flags.StringArrayVarP(&a.StringArray, "string-array", "", a.StringArray, "string array usage")

	err := Load(cfg, cmd, a)
	require.NoError(t, err)

	assert.Equal(t, true, a.Bool)
	assert.Equal(t, "stringValueEnv", a.String)
	assert.Equal(t, []string{"stringArrayValueEnv"}, a.StringArray)

	err = flags.Set("bool", "false")
	require.NoError(t, err)
	err = flags.Set("string", "stringValueFlag")
	require.NoError(t, err)
	err = flags.Set("string-array", "stringArrayValueFlag")
	require.NoError(t, err)

	err = Load(cfg, cmd, a)
	require.NoError(t, err)

	assert.Equal(t, false, a.Bool)
	assert.Equal(t, "stringValueFlag", a.String)
	assert.Equal(t, []string{"stringArrayValueFlag"}, a.StringArray)
}

func Test_homeDir(t *testing.T) {
	disableCache := homedir.DisableCache
	homedir.DisableCache = true
	t.Cleanup(func() {
		homedir.DisableCache = disableCache
	})

	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Setenv("HOME", path.Join(wd, "test-fixtures", "home-dir"))

	cmd, cfg, r, _ := setup(t)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "home-config-v", r.V)
}

func Test_xdgDir(t *testing.T) {
	t.Cleanup(func() {
		xdg.Reload()
	})

	disableCache := homedir.DisableCache
	homedir.DisableCache = true
	t.Cleanup(func() {
		homedir.DisableCache = disableCache
	})

	wd, err := os.Getwd()
	require.NoError(t, err)

	t.Setenv("HOME", path.Join(wd, "test-fixtures", "fake-home-dir"))
	t.Setenv("XDG_CONFIG_HOME", path.Join(wd, "test-fixtures", "xdg-dir"))

	xdg.Reload()

	cmd, cfg, r, _ := setup(t)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "xdg-config-v", r.V)
}

func Test_PostLoad(t *testing.T) {
	cfg := NewConfig("my-app")

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.ConfigFile = path.Join(wd, "test-fixtures", "config.yaml")

	r := &rootPostLoad{
		V: "default-v",
	}

	cmd := &cobra.Command{}

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "direct-config-v", r.V2)

	require.Equal(t, "direct-config-sub-v", r.Sub.Sv2)
}

type rootPostLoad struct {
	V   string `json:"v" yaml:"v" mapstructure:"v"`
	V2  string
	Sub subPostLoad `json:"sub" yaml:"sub" mapstructure:"sub"`
}

func (r *rootPostLoad) PostLoad() error {
	r.V2 = r.V
	return nil
}

var _ PostLoad = (*rootPostLoad)(nil)

type subPostLoad struct {
	Sv  string `json:"sv" yaml:"sv" mapstructure:"sv"`
	Sv2 string
}

func (s *subPostLoad) PostLoad() error {
	s.Sv2 = s.Sv
	return nil
}

var _ PostLoad = (*subPostLoad)(nil)
