package util

import (
	"fmt"
	"runtime/debug"
)

func ShowVersion() {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" {
		fmt.Println("Budva43 version:", info.Main.Version)
	} else {
		fmt.Println("Budva43 version: unknown")
	}
}
