package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToInt(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		want  int64
		panic bool
	}{
		{
			name: "valid case",
			s:    "123",
			want: 123,
		},
		{
			name:  "with dummy string",
			s:     "dummy",
			panic: true,
		},
		{
			name:  "with empty string",
			s:     "",
			panic: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.panic {
				assert.PanicsWithValue(t, fmt.Sprintf("ConvertToInt: strconv.Atoi: parsing \"%s\": invalid syntax", test.s), func() {
					ConvertToInt[int64](test.s)
				})
			} else {
				assert.Equal(t, test.want, ConvertToInt[int64](test.s))
			}
		})
	}
}

func TestCopy(t *testing.T) {
	type object struct {
		I int64
		A []string
		M map[string]string
	}
	obj1 := &object{}
	obj1.I = 1
	obj1.A = []string{"a", "b"}
	obj1.M = map[string]string{"a": "b"}
	obj2 := Copy(obj1)
	assert.Equal(t, obj2.I, int64(1))
	assert.Equal(t, obj2.A, []string{"a", "b"})
	assert.Equal(t, obj2.M, map[string]string{"a": "b"})
	obj2.I = 2
	obj2.A[0] = "c"
	obj2.M["a"] = "c"
	assert.Equal(t, obj1.I, int64(1), "obj1.I не изменился")
	assert.Equal(t, obj1.A, []string{"a", "b"}, "obj1.A не изменился")
	assert.Equal(t, obj1.M, map[string]string{"a": "b"}, "obj1.M не изменился")
}
