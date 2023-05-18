package fangs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/anchore/go-logger/adapter/discard"
)

func Test_Summarize(t *testing.T) {
	root := &cobra.Command{}

	cmd := &cobra.Command{}
	root.AddCommand(cmd)

	s0 := &summarize0{
		Name0: "name0 val",
		Type0: "type0 val",
		S2: summarize2{
			Field1: 10,
			Field2: true,
		},
	}

	s1 := &summarize1{
		Name: "s1 name",
		Type: "s1 type",
		S2: summarize2{
			Field1: 11,
		},
	}

	AddFlags(discard.New(), root.PersistentFlags(), s0)
	AddFlags(discard.New(), cmd.Flags(), s1)

	cfg := NewConfig("app")
	s := SummarizeCommand(cfg, cmd, s0, s1)
	require.Equal(t, `Name0: 'name0 val' # name0 usage flag (env: APP_NAME0)

Type0: 'type0 val' # type0 tag (env: APP_TYPE0)

s2-0:
  Field1: 10 # field 1 usage (env: APP_S2_0_FIELD1)
  
  Field2: true # field2 described (env: APP_S2_0_FIELD2)
  
Name: 's1 name' # described name (env: APP_NAME)

Type: 's1 type' # described type (env: APP_TYPE)

s2:
  Field1: 11 # field 1 usage (env: APP_S2_FIELD1)
  
  Field2: false # field2 described (env: APP_S2_FIELD2)`, s)
}

type summarize0 struct {
	Name0 string
	Type0 string     `description:"type0 tag"`
	S2    summarize2 `yaml:"s2-0"`
}

func (s *summarize0) AddFlags(flags FlagSet) {
	flags.StringVarP(&s.Name0, "name0", "", "name0 usage flag")
}

var _ FlagAdder = (*summarize0)(nil)

type summarize1 struct {
	Name string
	Type string     `description:"type description"`
	S2   summarize2 `yaml:"s2"`
}

func (s *summarize1) DescribeFields(d FieldDescriptionSet) {
	d.Add(&s.Name, "described name")
	d.Add(&s.Type, "described type")
}

func (s *summarize1) AddFlags(flags FlagSet) {
	flags.StringVarP(&s.Name, "name", "", "usage for name 1")
}

var _ FlagAdder = (*summarize1)(nil)
var _ FieldDescriber = (*summarize1)(nil)

type summarize2 struct {
	Field1 int
	Field2 bool
}

func (s *summarize2) DescribeFields(d FieldDescriptionSet) {
	d.Add(&s.Field2, "field2 described")
}

func (s *summarize2) AddFlags(flags FlagSet) {
	flags.IntVarP(&s.Field1, "field-1", "z", "field 1 usage")
}

var _ FlagAdder = (*summarize2)(nil)
var _ FieldDescriber = (*summarize2)(nil)

func Test_SummarizeValues(t *testing.T) {
	type TSub1 struct {
		Name string
		Val  int `yaml:"val-tsub1" description:"val1 inline tag description"`
	}
	type TSub2 struct {
		Name string `yaml:"name-tsub2"`
		Val  int    `description:"val2 inline tag description"`
	}
	type TSub3 struct {
		Name string `yaml:"name-tsub3"`
		Val  int
	}
	type T1 struct {
		TopBool   bool
		TopString string
		TSub1     `yaml:",inline,squash"`
		TSub2     `yaml:",inline"`
		TSub3     `yaml:"sub3"`
	}

	cfg := NewConfig("app")
	t1 := &T1{}

	cmd := &cobra.Command{}
	subCmd := &cobra.Command{}
	cmd.AddCommand(subCmd)

	cmd.Flags().StringVar(&t1.TopString, "top-string", "", "top-string command description")
	subCmd.Flags().StringVar(&t1.TSub2.Name, "sub2-name", "", "sub2-name command description")

	d1 := NewCommandDescriber(cfg.TagName, cmd)

	desc := NewDirectDescriber()
	desc.Add(&t1.TopBool, "top-bool manual description")
	desc.Add(&t1.TSub1.Name, "sub1-name manual description")
	desc.Add(&t1.TSub3.Val, "sub3-val manual description")

	describers := DescriptionProviders(d1, desc, NewStructDescriptionTagProvider())

	s := Summarize(cfg, describers, t1)

	require.Equal(t, `TopBool: false # top-bool manual description (env: APP_TOPBOOL)

TopString: '' # top-string command description (env: APP_TOPSTRING)

Name: '' # sub1-name manual description (env: APP_NAME)

val-tsub1: 0 # val1 inline tag description (env: APP_VAL_TSUB1)

TSub2:
  name-tsub2: '' # sub2-name command description (env: APP_TSUB2_NAME_TSUB2)
  
  Val: 0 # val2 inline tag description (env: APP_TSUB2_VAL)
  
sub3:
  name-tsub3: '' #  (env: APP_SUB3_NAME_TSUB3)
  
  Val: 0 # sub3-val manual description (env: APP_SUB3_VAL)`, s)
}

func Test_SummarizeLocations(t *testing.T) {
	t.Cleanup(func() {
		xdg.Reload()
	})

	disableCache := homedir.DisableCache
	homedir.DisableCache = true
	t.Cleanup(func() {
		homedir.DisableCache = disableCache
	})

	t.Setenv("HOME", "/home-dir")
	t.Setenv("XDG_CONFIG_HOME", "/xdg-home")
	t.Setenv("XDG_CONFIG_DIRS", "/xdg-dir1:/xdg-dir2")

	xdg.Reload()

	cfg := NewConfig("app")
	cfg.File = "/my-app/config.yaml"

	locations := SummarizeLocations(cfg)
	got := strings.Join(locations, "\n")

	allExts := func(path string) (out []string) {
		for _, ext := range viper.SupportedExts {
			out = append(out, path+"."+ext)
		}
		return
	}

	opts := []any{
		"/my-app/config.yaml",
		strings.Join(allExts(".app"), "\n"),
		strings.Join(allExts(".app/config"), "\n"),
		strings.Join(allExts("/home-dir/.app"), "\n"),
		strings.Join(allExts("/xdg-home/app/config"), "\n"),
		strings.Join(allExts("/xdg-dir1/app/config"), "\n"),
		strings.Join(allExts("/xdg-dir2/app/config"), "\n"),
	}

	expected := fmt.Sprintf(strings.Repeat("%s\n", len(opts)), opts...)

	require.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(got))
}
