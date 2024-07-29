package fangs

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_isPromotedMethod(t *testing.T) {
	s1 := &Sub2{}
	require.True(t, !isPromotedMethod(reflect.ValueOf(s1), "AddFlags"))

	type Ty1 struct {
		Something string
		Sub2
	}

	t1 := &Ty1{}
	require.True(t, isPromotedMethod(reflect.ValueOf(t1), "AddFlags"))

	type Ty2 struct {
		Ty1
	}

	t2 := &Ty2{}
	require.True(t, isPromotedMethod(reflect.ValueOf(t2), "AddFlags"))

	// reflect-created structs do not include promoted methods
	tt1 := reflect.TypeOf(t1)
	f := tt1.Elem().Field(1)
	ty3 := reflect.StructOf([]reflect.StructField{f})
	t3 := reflect.New(ty3).Interface()
	_, ok := ty3.MethodByName("AddFlags")

	assert.False(t, ok)
	// not a promoted method because the method doesn't exist on the struct
	require.True(t, !isPromotedMethod(reflect.ValueOf(t3), "AddFlags"))
}
