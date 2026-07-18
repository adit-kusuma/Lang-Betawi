//go:build windows

package main

import (
	"fmt"
	"os"

	"language-betawi/internal/installer"
)

func main() {
	if err := installer.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
