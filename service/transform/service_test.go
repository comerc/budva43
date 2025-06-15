package transform

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeMarkdown(t *testing.T) {
	t.Parallel()

	s1 := "_ * ( ) ~ ` > # + = | { } . !"
	s2 := `\[ \] \-`
	a := strings.Split(s1+" "+s2, " ")
	for _, v := range a {
		assert.Equal(t, `\`+v, escapeMarkdown(v))
	}
}
