package fangs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Indent(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		indent string
		want   string
	}{
		{
			name:   "no indent",
			text:   "single line",
			indent: "",
			want:   "single line",
		},
		{
			name:   "single line",
			text:   "single line",
			indent: "  ",
			want:   "  single line",
		},
		{
			name:   "multi line",
			text:   "multi\nline",
			indent: "  ",
			want:   "  multi\n  line",
		},
		{
			name:   "keep trailing newline",
			text:   "multi\nline\n\n",
			indent: "  ",
			want:   "  multi\n  line\n  \n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Indent(tt.text, tt.indent))
		})
	}
}

func Test_isPromotedMethod(t *testing.T) {
	s1 := &Sub2{}
	require.True(t, !isPromotedMethod(s1, "AddFlags"))

	type Ty1 struct {
		Something string
		Sub2
	}

	t1 := &Ty1{}
	require.True(t, isPromotedMethod(t1, "AddFlags"))

	type Ty2 struct {
		Ty1
	}

	t2 := &Ty2{}
	require.True(t, isPromotedMethod(t2, "AddFlags"))

	// reflect-created structs do not include promoted methods
	tt1 := reflect.TypeOf(t1)
	f := tt1.Elem().Field(1)
	ty3 := reflect.StructOf([]reflect.StructField{f})
	t3 := reflect.New(ty3).Interface()
	_, ok := ty3.MethodByName("AddFlags")

	assert.False(t, ok)
	// not a promoted method because the method doesn't exist on the struct
	require.True(t, !isPromotedMethod(t3, "AddFlags"))
}
