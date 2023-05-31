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

	type ty0 struct {
		S0 summarize0 `mapstructure:",squash"`
		S1 summarize1 `mapstructure:",squash"`
	}

	t0 := &ty0{
		S0: summarize0{
			Name0: "name0 val",
			Type0: "type0 val",
			S2: summarize2{
				Field1: 10,
				Field2: true,
			},
		},
		S1: summarize1{
			Name: "s1 name",
			Type: "s1 type",
			S2: summarize2{
				Field1: 11,
			},
		},
	}

	AddFlags(discard.New(), root.PersistentFlags(), &t0.S0)
	AddFlags(discard.New(), cmd.Flags(), &t0.S1)

	cfg := NewConfig("app")
	s := SummarizeCommand(cfg, cmd, t0, &t0.S0, &t0.S1)
	require.Equal(t, `# name0 usage flag (env: APP_NAME0)
Name0: 'name0 val'

# type0 tag (env: APP_TYPE0)
Type0: 'type0 val'

s2-0:
  # field 1 usage (env: APP_S2_0_FIELD1)
  Field1: 10
  
  # field2 described
  # multiline (env: APP_S2_0_FIELD2)
  Field2: true
  
# described name (env: APP_NAME)
Name: 's1 name'

# described type (env: APP_TYPE)
Type: 's1 type'

s2:
  # field 1 usage (env: APP_S2_FIELD1)
  Field1: 11
  
  # field2 described
  # multiline (env: APP_S2_FIELD2)
  Field2: false
  
`, s)
}

type summarize0 struct {
	Name0      string
	Type0      string     `description:"type0 tag"`
	S2         summarize2 `mapstructure:"s2-0"`
	unexported summarize2
}

func (s *summarize0) AddFlags(flags FlagSet) {
	flags.StringVarP(&s.Name0, "name0", "", "name0 usage flag")
}

var _ FlagAdder = (*summarize0)(nil)

type summarize1 struct {
	Name string
	Type string     `description:"type description"`
	S2   summarize2 `mapstructure:"s2"`
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
	d.Add(&s.Field2, "field2 described\nmultiline")
}

func (s *summarize2) AddFlags(flags FlagSet) {
	flags.IntVarP(&s.Field1, "field-1", "z", "field 1 usage")
}

var _ FlagAdder = (*summarize2)(nil)
var _ FieldDescriber = (*summarize2)(nil)

func Test_SummarizeValues(t *testing.T) {
	type TSub1 struct {
		Name string
		Val  int `mapstructure:"val-tsub1" description:"val1 inline tag description"`
	}
	type TSub2 struct {
		Name string `mapstructure:"name-tsub2"`
		Val  int    `description:"val2 inline tag description"`
	}
	type TSub3 struct {
		Name string `mapstructure:"name-tsub3"`
		Val  int
	}
	type TSub4 struct {
		TSub1
		Sub2 TSub2
	}
	type T1 struct {
		TopBool   bool
		TopString string
		TSub1     `mapstructure:",squash"`
		TSub2     `mapstructure:""`
		TSub3     `mapstructure:"sub3"`
		TSub4     `mapstructure:"sub4"`
	}

	cfg := NewConfig("app")
	t1 := &T1{}

	cmd := &cobra.Command{}
	subCmd := &cobra.Command{}
	cmd.AddCommand(subCmd)

	cmd.Flags().StringVar(&t1.TopString, "top-string", "", "top-string command description")
	subCmd.Flags().StringVar(&t1.TSub2.Name, "sub2-name", "", "sub2-name command description")

	d1 := NewCommandFlagDescriptionProvider(cfg.TagName, cmd)

	desc := NewDirectDescriber()
	desc.Add(&t1.TopBool, "top-bool manual description")
	desc.Add(&t1.TSub1.Name, "sub1-name manual description")
	desc.Add(&t1.TSub3.Val, "sub3-val manual description")

	describers := DescriptionProviders(d1, desc, NewStructDescriptionTagProvider())

	s := Summarize(cfg, describers, t1)

	require.Equal(t, `# top-bool manual description (env: APP_TOPBOOL)
TopBool: false

# top-string command description (env: APP_TOPSTRING)
TopString: ''

# sub1-name manual description (env: APP_NAME)
Name: ''

# val1 inline tag description (env: APP_VAL_TSUB1)
val-tsub1: 0

TSub2:
  # sub2-name command description (env: APP_TSUB2_NAME_TSUB2)
  name-tsub2: ''
  
  # val2 inline tag description (env: APP_TSUB2_VAL)
  Val: 0
  
sub3:
  # (env: APP_SUB3_NAME_TSUB3)
  name-tsub3: ''
  
  # sub3-val manual description (env: APP_SUB3_VAL)
  Val: 0
  
sub4:
  TSub1:
    # (env: APP_SUB4_TSUB1_NAME)
    Name: ''
    
    # val1 inline tag description (env: APP_SUB4_TSUB1_VAL_TSUB1)
    val-tsub1: 0
    
  Sub2:
    # (env: APP_SUB4_SUB2_NAME_TSUB2)
    name-tsub2: ''
    
    # val2 inline tag description (env: APP_SUB4_SUB2_VAL)
    Val: 0
    
`, s)
}

type Summarize1 struct {
	Name string
	Val  int `mapstructure:"summarize1-val" description:"summarize1-val inline tag description"`
}

type Summarize2 struct {
	Name string `mapstructure:"summarize2-name"`
	Val  int
}

func (s *Summarize2) DescribeFields(d FieldDescriptionSet) {
	d.Add(&s.Val, "val 2 description")
}

func (s *Summarize2) AddFlags(flags FlagSet) {
	flags.StringVarP(&s.Name, "summarize2-name", "", "summarize2-name command description")
}

var _ FlagAdder = (*Summarize2)(nil)
var _ FieldDescriber = (*Summarize2)(nil)

func Test_SummarizeValuesWithPointers(t *testing.T) {
	type T1 struct {
		TopBool    bool
		TopString  string
		Summarize1 `mapstructure:",squash"`
		Pointer    *Summarize2 `mapstructure:"ptr"`
		NilPointer *Summarize2 `mapstructure:"nil"`
	}

	cfg := NewConfig("my-app")
	t1 := &T1{
		Pointer: &Summarize2{
			Name: "summarize2 name",
			Val:  2,
		},
	}

	cmd := &cobra.Command{}
	subCmd := &cobra.Command{}
	cmd.AddCommand(subCmd)

	cmd.Flags().StringVar(&t1.TopString, "top-string", "", "top-string command description")
	AddFlags(cfg.Logger, subCmd.Flags(), t1)

	s := SummarizeCommand(cfg, subCmd, t1)

	require.Equal(t, `# (env: MY_APP_TOPBOOL)
TopBool: false

# top-string command description (env: MY_APP_TOPSTRING)
TopString: ''

# (env: MY_APP_NAME)
Name: ''

# summarize1-val inline tag description (env: MY_APP_SUMMARIZE1_VAL)
summarize1-val: 0

ptr:
  # summarize2-name command description (env: MY_APP_PTR_SUMMARIZE2_NAME)
  summarize2-name: 'summarize2 name'
  
  # val 2 description (env: MY_APP_PTR_VAL)
  Val: 2
  
nil:
  # (env: MY_APP_NIL_SUMMARIZE2_NAME)
  summarize2-name: ''
  
  # val 2 description (env: MY_APP_NIL_VAL)
  Val: 0
  
`, s)
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
