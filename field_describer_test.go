package fangs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_fieldDescriber(t *testing.T) {
	f1 := &fdTest1{}

	d := NewFieldDescriber(f1)

	require.Equal(t, 1, f1.called)
	require.Equal(t, 1, f1.FdTest2.called)

	dd := d.(*directDescriber)

	var values []string
	for _, d := range dd.flagRefs {
		values = append(values, d.Usage)
	}

	require.Contains(t, values, "string description")
	require.Contains(t, values, "int description")
}

type fdTest1 struct {
	called  int
	String  string
	FdTest2 fdTest2
}

func (f *fdTest1) DescribeFields(d FieldDescriptionSet) {
	f.called++
	d.Add(&f.String, "string description")
}

var _ FieldDescriber = &fdTest1{}

type fdTest2 struct {
	called int
	Int    int
}

func (f *fdTest2) DescribeFields(d FieldDescriptionSet) {
	f.called++
	d.Add(&f.Int, "int description")
}

var _ FieldDescriber = &fdTest2{}
