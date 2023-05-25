package fangs

import (
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sub struct {
	Sv      string `mapstructure:"sv"`
	Unbound string `mapstructure:"unbound"`
}

type root struct {
	V   string `mapstructure:"v"`
	Sub *sub   `mapstructure:"sub"`
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
	cfg.File = path.Join(wd, "test-fixtures", "config.yaml")

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "direct-config-sub-v", s.Sv)
	require.Equal(t, "direct-config-v", r.V)
}

func Test_LoadEmbeddedSquash(t *testing.T) {
	type Top struct {
		Value string
	}
	type Sub struct {
		Value string
	}
	type t1 struct {
		Top `mapstructure:",squash"`
		Sub `mapstructure:"sub2"`
	}

	cfg := NewConfig("app")
	cmd := &cobra.Command{}

	v := &t1{
		Top: Top{},
		Sub: Sub{},
	}

	t.Setenv("APP_VALUE", "top-v")
	t.Setenv("APP_SUB2_VALUE", "sub2-v")

	err := Load(cfg, cmd, v)
	require.NoError(t, err)

	require.Equal(t, "top-v", v.Top.Value)
	require.Equal(t, "sub2-v", v.Sub.Value)
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

func Test_LoadFromEnvOnly(t *testing.T) {
	t.Setenv("APP_V", "env-var-v")
	t.Setenv("APP_SUB_SV", "env-var-sv")

	cmd := &cobra.Command{}
	s := &sub{
		Sv: "default-sv",
	}
	r := &root{
		V:   "default-v",
		Sub: s,
	}

	cfg := NewConfig("app")

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
	cfg.File = path.Join(wd, "test-fixtures", "config.yaml")

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
	cfg.File = path.Join(wd, "test-fixtures", "config.yaml")

	err = LoadAt(cfg, cmd, "sub", s)
	require.NoError(t, err)

	require.Equal(t, "env-var-sv", s.Sv)
}

func Test_LoadSubStructEnv(t *testing.T) {
	cmd, cfg, _, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.File = path.Join(wd, "test-fixtures", "config.yaml")

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
	cfg.File = path.Join(wd, "test-fixtures", "config.yaml")

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

func p[T any](t T) *T {
	return &t
}

func Test_flagBoolPtrValues(t *testing.T) {
	type s struct {
		Bool *bool `mapstructure:"bool"`
	}
	a := &s{}

	cmd := &cobra.Command{}
	flags := cmd.Flags()
	BoolPtrVarP(flags, &a.Bool, "bool", "", "")

	refs := commandFlagRefs(cmd)
	require.NotEmpty(t, refs)

	err := flags.Set("bool", "true")
	require.NoError(t, err)
	require.NotNil(t, a.Bool)
	require.Equal(t, true, *a.Bool)

	t.Setenv("APP_BOOL", "false")

	cfg := NewConfig("app")
	err = Load(cfg, cmd, a)
	require.NoError(t, err)
	require.NotNil(t, a.Bool)
	require.Equal(t, true, *a.Bool)
}

func Test_AllFieldTypes(t *testing.T) {
	appName := "app"
	envName := func(name string) string {
		name = appName + "." + name
		name = regexp.MustCompile("[^a-zA-Z0-9]").ReplaceAllString(name, "_")
		return strings.ToUpper(name)
	}

	type all struct {
		Bool        bool     `mapstructure:"bool"`
		BoolPtr     *bool    `mapstructure:"bool-ptr"`
		Int         int      `mapstructure:"int"`
		IntPtr      *int     `mapstructure:"int-ptr"`
		String      string   `mapstructure:"string"`
		StringPtr   *string  `mapstructure:"string-ptr"`
		StringArray []string `mapstructure:"string-array"`
	}

	tests := []struct {
		name     string
		env      map[string]string
		flags    map[string]string
		expected *all
	}{
		{
			name: "all values from env",
			// NOTE this test needs to include all the env vars -- the names are used to reset the env vars
			env: map[string]string{
				"bool":         "true",
				"bool-ptr":     "false",
				"int":          "8",
				"int-ptr":      "9",
				"string":       "stringValueEnv",
				"string-ptr":   "stringValuePtrEnv",
				"string-array": "stringArrayValueEnv",
			},
			expected: &all{
				Bool:        true,
				BoolPtr:     p(false),
				Int:         8,
				IntPtr:      p(9),
				String:      "stringValueEnv",
				StringPtr:   p("stringValuePtrEnv"),
				StringArray: []string{"stringArrayValueEnv"},
			},
		},
		{
			name: "all values from config",
			expected: &all{
				Bool:        false,
				BoolPtr:     p(true),
				Int:         2,
				IntPtr:      p(3),
				String:      "stringValueConfig",
				StringPtr:   p("stringValuePtrConfig"),
				StringArray: []string{"stringArrayValueConfig"},
			},
		},
		{
			name: "all values from flags",
			env: map[string]string{
				"bool":         "true",
				"bool-ptr":     "false",
				"int":          "8",
				"int-ptr":      "9",
				"string":       "stringValueEnv",
				"string-ptr":   "stringValuePtrEnv",
				"string-array": "stringArrayValueEnv",
			},
			flags: map[string]string{
				"bool":         "false",
				"bool-ptr":     "true",
				"int":          "5",
				"string":       "stringValueFlag",
				"string-array": "stringArrayValueFlag",
			},
			expected: &all{
				Bool:        false,
				BoolPtr:     p(true),
				Int:         5,
				IntPtr:      p(9),
				String:      "stringValueFlag",
				StringPtr:   p("stringValuePtrEnv"),
				StringArray: []string{"stringArrayValueFlag"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if len(test.env) > 0 {
				// reset all the env vars -- use the :
				for k := range tests[0].env {
					t.Setenv(envName(k), "")
					_ = os.Unsetenv(envName(k))
				}

				// set for the test
				for k, v := range test.env {
					t.Setenv(envName(k), v)
				}
			}

			cfg := NewConfig(appName)
			cfg.File = "test-fixtures/all-values/app.yaml"

			cmd := &cobra.Command{}

			a := &all{}

			flags := cmd.Flags()
			flags.BoolVarP(&a.Bool, "bool", "", a.Bool, "bool usage")
			BoolPtrVarP(flags, &a.BoolPtr, "bool-ptr", "", "bool ptr usage")
			flags.IntVarP(&a.Int, "int", "", a.Int, "int usage")
			IntPtrVarP(flags, &a.IntPtr, "int-ptr", "", "int ptr usage")
			flags.StringVarP(&a.String, "string", "", a.String, "string usage")
			StringPtrVarP(flags, &a.StringPtr, "string-ptr", "", "string ptr usage")
			flags.StringArrayVarP(&a.StringArray, "string-array", "", a.StringArray, "string array usage")

			for k, v := range test.flags {
				err := flags.Set(k, v)
				require.NoError(t, err)
			}

			err := Load(cfg, cmd, a)
			require.NoError(t, err)

			assert.Equal(t, test.expected, a)
		})
	}
}

func Test_wdConfigYaml(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(path.Join(wd, "test-fixtures", "wd-config"))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", path.Join(wd, "test-fixtures", "fake-home-dir"))

	cmd, cfg, r, _ := setup(t)

	cfg.Finders = append(cfg.Finders, FindConfigYamlInCwd)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "wd-config-v", r.V)
}

func Test_wdSubdirConfigYaml(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	err = os.Chdir(path.Join(wd, "test-fixtures", "wd-subdir"))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	t.Setenv("HOME", path.Join(wd, "test-fixtures", "fake-home-dir"))

	cmd, cfg, r, _ := setup(t)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "wd-subdir-config-v", r.V)
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

func Test_xdgDirs(t *testing.T) {
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
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_DIRS", path.Join(wd, "test-fixtures", "xdg-dir"))

	xdg.Reload()

	cmd, cfg, r, _ := setup(t)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "xdg-config-v", r.V)
}

func Test_xdgHomeDir(t *testing.T) {
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
	t.Setenv("XDG_CONFIG_HOME", path.Join(wd, "test-fixtures", "xdg-home"))
	t.Setenv("XDG_CONFIG_DIRS", path.Join(wd, "test-fixtures", "xdg-dir"))

	xdg.Reload()

	cmd, cfg, r, _ := setup(t)

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "xdg-home-config-v", r.V)
}

func Test_NilPointerFields(t *testing.T) {
	cfg := NewConfig("my-app")

	r := &rootPostLoad{}

	cmd := &cobra.Command{}

	err := Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Nil(t, r.Bool)

	t.Setenv("MY_APP_BOOL", "true")
	t.Setenv("MY_APP_PTR_SV", "env-sv")

	r = &rootPostLoad{}

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.NotNil(t, r.Bool)
	require.True(t, *r.Bool)

	require.NotNil(t, r.Ptr)
	require.Equal(t, "env-sv", r.Ptr.Sv)
}

func Test_PostLoad(t *testing.T) {
	cfg := NewConfig("my-app")

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.File = path.Join(wd, "test-fixtures", "config.yaml")

	r := &rootPostLoad{
		V: "default-v",
		Ptr: &subPostLoad{
			Sv: "ptr-v",
		},
	}

	cmd := &cobra.Command{}

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "direct-config-v", r.V2)
	require.Equal(t, "direct-config-sub-v", r.Sub.Sv2)
	require.Equal(t, "direct-config-sub-sub-v", r.Sub.Sub2.Ssv2)
	require.Equal(t, "direct-config-sub-sub-sub-v", r.Sub.Sub2.Sub3.Sssv2)
	require.Equal(t, "ptr-v", r.Ptr.Sv2)
}

type rootPostLoad struct {
	V    string `mapstructure:"v"`
	V2   string
	Bool *bool        `mapstructure:"bool"`
	Ptr  *subPostLoad `mapstructure:"ptr"`
	Sub  subPostLoad  `mapstructure:"sub"`
}

func (r *rootPostLoad) PostLoad() error {
	r.V2 = r.V
	return nil
}

var _ PostLoad = (*rootPostLoad)(nil)

type subPostLoad struct {
	Sv   string `mapstructure:"sv"`
	Sv2  string
	Sub2 subSubPostLoad `mapstructure:"sub2"`
}

func (s *subPostLoad) PostLoad() error {
	s.Sv2 = s.Sv
	return nil
}

var _ PostLoad = (*subPostLoad)(nil)

type subSubPostLoad struct {
	Ssv  string `mapstructure:"ssv"`
	Ssv2 string
	Sub3 subSubSubPostLoad `mapstructure:"sub3"`
}

func (s *subSubPostLoad) PostLoad() error {
	s.Ssv2 = s.Ssv
	return nil
}

var _ PostLoad = (*subSubPostLoad)(nil)

type subSubSubPostLoad struct {
	Sssv  string `mapstructure:"sssv"`
	Sssv2 string
}

func (s *subSubSubPostLoad) PostLoad() error {
	s.Sssv2 = s.Sssv
	return nil
}

var _ PostLoad = (*subSubSubPostLoad)(nil)
