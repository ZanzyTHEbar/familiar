//go:build !linux

package task

import "io"

func Run(_ []string, _ io.Reader, _ io.Writer, stderr io.Writer) int {
	io.WriteString(stderr, "familiar task requires Linux openat2; no weaker filesystem fallback\n")
	return 2
}
