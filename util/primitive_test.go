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
