package util

import (
	"fmt"
	"os"
	"strings"
)

func GetFlag(name string) *string {
	prefix := fmt.Sprintf("-%s=", name) // только для флагов через "="
	var result *string
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			v := arg[len(prefix):]
			result = &v
		}
	}
	return result
}

func HasFlag(name string) bool {
	prefix := fmt.Sprintf("-%s", name)
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			return true
		}
	}
	return false
}
