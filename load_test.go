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
	cfg.Files = []string{path.Join(wd, "test-fixtures", "config.yaml")}

	err = Load(cfg, cmd, r)
	require.NoError(t, err)

	require.Equal(t, "direct-config-sub-v", s.Sv)
	require.Equal(t, "direct-config-v", r.V)
}

func Test_Multilevel(t *testing.T) {
	cmd, cfg, r, _ := setup(t)

	cfg.Files = []string{"test-fixtures/multilevel/1.yaml", "test-fixtures/multilevel/2.yaml"}

	err := Load(cfg, cmd, r)
	require.NoError(t, err)
	require.Equal(t, "level-2", r.Sub.Sv)
}

func Test_Profile(t *testing.T) {
	cmd, cfg, r, _ := setup(t)

	cfg.Profiles = []string{"override"}
	cfg.Files = []string{"test-fixtures/multilevel/2.yaml", "test-fixtures/multilevel/1.yaml"}

	err := Load(cfg, cmd, r)
	require.NoError(t, err)
	require.Equal(t, "level-override", r.Sub.Sv)
}

func Test_MultilevelSlices(t *testing.T) {
	cmd, cfg, _, _ := setup(t)

	type slice struct {
		StringArray []string `mapstructure:"string-array"`
	}
	type holder struct {
		Slice []slice `mapstructure:"slice"`
	}

	r := &holder{}

	cfg.Files = []string{"test-fixtures/multilevel/1.yaml", "test-fixtures/multilevel/2.yaml"}

	err := Load(cfg, cmd, r)
	require.NoError(t, err)
	require.Len(t, r.Slice, 2)
	require.Equal(t, "v1.1", r.Slice[0].StringArray[0])
	require.Equal(t, "v2.1", r.Slice[1].StringArray[0])
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
	cfg.Files = []string{path.Join(wd, "test-fixtures", "config.yaml")}

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
	cfg.Files = []string{path.Join(wd, "test-fixtures", "config.yaml")}

	err = LoadAt(cfg, cmd, "sub", s)
	require.NoError(t, err)

	require.Equal(t, "env-var-sv", s.Sv)
}

func Test_LoadSubStructEnv(t *testing.T) {
	cmd, cfg, _, s := setup(t)

	wd, err := os.Getwd()
	require.NoError(t, err)
	cfg.Files = []string{path.Join(wd, "test-fixtures", "config.yaml")}

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
	cfg.Files = []string{path.Join(wd, "test-fixtures", "config.yaml")}

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

func Test_zeroFields(t *testing.T) {
	type s struct {
		List []string `mapstructure:"list"`
	}
	a := &s{
		List: []string{
			"default1",
			"default2",
			"default3",
		},
	}

	cmd := &cobra.Command{}

	cfg := NewConfig("app")
	err := Load(cfg, cmd, a)
	require.NoError(t, err)

	require.Equal(t, []string{"default1", "default2", "default3"}, a.List)

	t.Setenv("APP_LIST", "set1,set2")

	cfg = NewConfig("app")
	err = Load(cfg, cmd, a)
	require.NoError(t, err)

	require.Equal(t, []string{"set1", "set2"}, a.List)
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
		Float64     float64  `mapstructure:"float64"`
		Float64Ptr  *float64 `mapstructure:"float64-ptr"`
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
				"float64":      "3.14",
				"float64-ptr":  "2.718",
			},
			expected: &all{
				Bool:        true,
				BoolPtr:     p(false),
				Int:         8,
				IntPtr:      p(9),
				String:      "stringValueEnv",
				StringPtr:   p("stringValuePtrEnv"),
				StringArray: []string{"stringArrayValueEnv"},
				Float64:     3.14,
				Float64Ptr:  p(2.718),
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
				Float64:     1.618,
				Float64Ptr:  p(0.618),
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
				"float64":      "4.44",
				"float64-ptr":  "5.55",
			},
			flags: map[string]string{
				"bool":         "false",
				"bool-ptr":     "true",
				"int":          "5",
				"string":       "stringValueFlag",
				"string-array": "stringArrayValueFlag",
				"float64":      "3.14",
				"float64-ptr":  "2.718",
			},
			expected: &all{
				Bool:        false,
				BoolPtr:     p(true),
				Int:         5,
				IntPtr:      p(9),
				String:      "stringValueFlag",
				StringPtr:   p("stringValuePtrEnv"),
				StringArray: []string{"stringArrayValueFlag"},
				Float64:     3.14,
				Float64Ptr:  p(2.718),
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
			cfg.Files = []string{"test-fixtures/all-values/app.yaml"}

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
			flags.Float64VarP(&a.Float64, "float64", "", a.Float64, "float64 usage")
			Float64PtrVarP(flags, &a.Float64Ptr, "float64-ptr", "", "float64 ptr usage")

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
	cfg.Files = []string{path.Join(wd, "test-fixtures", "config.yaml")}

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

type Public struct {
	Value bool `json:"value" yaml:"value" mapstructure:"value"`
}

type private struct {
	Value bool `json:"value" yaml:"value" mapstructure:"value"`
}

func Test_EmbeddedPublicStruct(t *testing.T) {
	val := &struct {
		Public      `yaml:",inline" mapstructure:",squash"`
		PublicField struct {
			Public `yaml:",inline" mapstructure:",squash"`
		} `yaml:"field" mapstructure:"field"`
	}{}

	cfg := NewConfig("app")
	t.Setenv("APP_VALUE", "true")
	t.Setenv("APP_FIELD_VALUE", "true")

	cmd := &cobra.Command{}
	err := Load(cfg, cmd, val)

	require.NoError(t, err)
	require.NotNil(t, val.Public)
	require.True(t, val.Public.Value)
	require.NotNil(t, val.PublicField.Public)
	require.True(t, val.PublicField.Public.Value)
}

func Test_EmbeddedPublicStructPointer(t *testing.T) {
	val := &struct {
		*Public     `yaml:",inline" mapstructure:",squash"`
		PublicField struct {
			*Public `yaml:",inline" mapstructure:",squash"`
		} `yaml:"field" mapstructure:"field"`
	}{}

	cfg := NewConfig("app")
	t.Setenv("APP_VALUE", "true")
	t.Setenv("APP_FIELD_VALUE", "true")

	cmd := &cobra.Command{}
	err := Load(cfg, cmd, val)

	require.NoError(t, err)
	require.NotNil(t, val.Public)
	require.True(t, val.Public.Value)
	require.NotNil(t, val.PublicField.Public)
	require.True(t, val.PublicField.Value)
}

func Test_EmbeddedPrivateStruct(t *testing.T) {
	val := &struct {
		private     `yaml:",inline" mapstructure:",squash"`
		PublicField struct {
			private `yaml:",inline" mapstructure:",squash"`
		} `yaml:"field" mapstructure:"field"`
	}{}

	cfg := NewConfig("app")
	t.Setenv("APP_VALUE", "true")
	t.Setenv("APP_FIELD_VALUE", "true")

	cmd := &cobra.Command{}
	err := Load(cfg, cmd, val)

	require.NoError(t, err)
	require.NotNil(t, val.private)
	require.True(t, val.private.Value)
	require.NotNil(t, val.PublicField.private)
	require.True(t, val.PublicField.private.Value)
}

func Test_EmbeddedPrivateStructPointer(t *testing.T) {
	// Note that, unlike Test_EmbeddedPublicStructPointer above,
	// in this test, *private is not exported and cannot be set or addressed.
	// This is a language limitation. See https://go-review.googlesource.com/c/go/+/53643
	// and https://github.com/golang/go/issues/21357
	val := &struct {
		*private    `yaml:",inline" mapstructure:",squash"`
		PublicField struct {
			*private `yaml:",inline" mapstructure:",squash"`
		} `yaml:"field" mapstructure:"field"`
	}{}

	cfg := NewConfig("app")
	t.Setenv("APP_VALUE", "true")
	t.Setenv("APP_FIELD_VALUE", "true")

	cmd := &cobra.Command{}
	err := Load(cfg, cmd, val)

	// https://github.com/mitchellh/mapstructure/blob/bf980b35cac4dfd34e05254ee5aba086504c3f96/mapstructure.go#L1338
	assert.ErrorContains(t, err, "unsupported type for squash")
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

var _ PostLoader = (*rootPostLoad)(nil)

type subPostLoad struct {
	Sv   string `mapstructure:"sv"`
	Sv2  string
	Sub2 subSubPostLoad `mapstructure:"sub2"`
}

func (s *subPostLoad) PostLoad() error {
	s.Sv2 = s.Sv
	return nil
}

var _ PostLoader = (*subPostLoad)(nil)

type subSubPostLoad struct {
	Ssv  string `mapstructure:"ssv"`
	Ssv2 string
	Sub3 subSubSubPostLoad `mapstructure:"sub3"`
}

func (s *subSubPostLoad) PostLoad() error {
	s.Ssv2 = s.Ssv
	return nil
}

var _ PostLoader = (*subSubPostLoad)(nil)

type subSubSubPostLoad struct {
	Sssv  string `mapstructure:"sssv"`
	Sssv2 string
}

func (s *subSubSubPostLoad) PostLoad() error {
	s.Sssv2 = s.Sssv
	return nil
}

var _ PostLoader = (*subSubSubPostLoad)(nil)

func Test_postLoadSlices(t *testing.T) {
	type typ struct {
		Items       []item
		ItemPtrs    []*item
		PtrItems    *[]item
		PtrItemPtrs *[]*item
	}

	type top struct {
		Sub typ
	}

	v := top{
		Sub: typ{
			Items: []item{
				{V: "1"},
				{V: "2"},
			},
			ItemPtrs: []*item{
				{V: "1"},
				{V: "2"},
			},
			PtrItems: &[]item{
				{V: "1"},
				{V: "2"},
			},
			PtrItemPtrs: &[]*item{
				{V: "1"},
				{V: "2"},
			},
		},
	}

	err := Load(NewConfig("app"), &cobra.Command{}, &v)
	require.NoError(t, err)

	tested := 0
	for _, i := range v.Sub.Items {
		tested++
		assert.NotEmpty(t, i.loadedValue, "Items")
		assert.Equalf(t, i.V, i.loadedValue, "Items")
	}
	for _, i := range v.Sub.ItemPtrs {
		tested++
		assert.NotEmpty(t, i.loadedValue, "ItemPtrs")
		assert.Equalf(t, i.V, i.loadedValue, "ItemPtrs")
	}
	for _, i := range *v.Sub.PtrItems {
		tested++
		assert.NotEmpty(t, i.loadedValue, "PtrItems")
		assert.Equalf(t, i.V, i.loadedValue, "PtrItems")
	}
	for _, i := range *v.Sub.PtrItemPtrs {
		tested++
		assert.NotEmpty(t, i.loadedValue, "PtrItemPtrs")
		assert.Equalf(t, i.V, i.loadedValue, "PtrItemPtrs")
	}
	require.Equal(t, 8, tested)
}

func Test_postLoadMaps(t *testing.T) {
	type typ struct {
		Items       map[int]item
		ItemPtrs    map[int]*item
		PtrItems    *map[int]item
		PtrItemPtrs *map[int]*item
	}

	type top struct {
		Sub typ
	}

	v := top{
		Sub: typ{
			Items: map[int]item{
				1: {V: "1"},
				2: {V: "2"},
			},
			ItemPtrs: map[int]*item{
				1: {V: "1"},
				2: {V: "2"},
			},
			PtrItems: &map[int]item{
				1: {V: "1"},
				2: {V: "2"},
			},
			PtrItemPtrs: &map[int]*item{
				1: {V: "1"},
				2: {V: "2"},
			},
		},
	}

	err := Load(NewConfig("app"), &cobra.Command{}, &v)
	require.NoError(t, err)

	tested := 0
	for _, i := range v.Sub.Items {
		tested++
		assert.NotEmpty(t, i.loadedValue, "Items")
		assert.Equalf(t, i.V, i.loadedValue, "Items")
	}
	for _, i := range v.Sub.ItemPtrs {
		tested++
		assert.NotEmpty(t, i.loadedValue, "ItemPtrs")
		assert.Equalf(t, i.V, i.loadedValue, "ItemPtrs")
	}
	for _, i := range *v.Sub.PtrItems {
		tested++
		assert.NotEmpty(t, i.loadedValue, "PtrItems")
		assert.Equalf(t, i.V, i.loadedValue, "PtrItems")
	}
	for _, i := range *v.Sub.PtrItemPtrs {
		tested++
		assert.NotEmpty(t, i.loadedValue, "PtrItemPtrs")
		assert.Equalf(t, i.V, i.loadedValue, "PtrItemPtrs")
	}
	require.Equal(t, 8, tested)
}

type item struct {
	loadedValue string
	V           string
}

var _ PostLoader = (*item)(nil)

func (s *item) PostLoad() error {
	s.loadedValue = s.V
	return nil
}
