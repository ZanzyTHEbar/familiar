//go:build !linux

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "scoped search is supported only on Linux")
	os.Exit(1)
}
