package fangs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_Summarize(t *testing.T) {
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
	cmd.Flags().StringVar(&t1.TopString, "top-string", "", "top-string command description")
	cmd.Flags().StringVar(&t1.TSub2.Name, "sub2-name", "", "sub2-name command description")

	d1 := NewCommandDescriptions(cfg, cmd)

	desc := NewDescribers()
	desc.Add(&t1.TopBool, "top-bool manual description")
	desc.Add(&t1.TSub1.Name, "sub1-name manual description")
	desc.Add(&t1.TSub3.Val, "sub3-val manual description")

	s := Summarize(cfg, t1, d1, desc)

	require.Equal(t, `TopBool: false # top-bool manual description (env var: APP_TOPBOOL)

TopString: '' # top-string command description (env var: APP_TOPSTRING)

Name: '' # sub1-name manual description (env var: APP_NAME)

val-tsub1: 0 # val1 inline tag description (env var: APP_VAL_TSUB1)

TSub2:
  name-tsub2: '' # sub2-name command description (env var: APP_TSUB2_NAME_TSUB2)
  
  Val: 0 # val2 inline tag description (env var: APP_TSUB2_VAL)
  
sub3:
  name-tsub3: '' #  (env var: APP_SUB3_NAME_TSUB3)
  
  Val: 0 # sub3-val manual description (env var: APP_SUB3_VAL)`, s)
}
