// Familiar's F3 CLI is a presentation layer over package doctor.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"familiar.local/familiar/doctor"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "doctor" {
		fmt.Fprintln(stderr, "usage: familiar doctor [--json] [--offline]")
		return 2
	}
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "emit schema-versioned JSON readiness")
	flags.Bool("offline", true, "offline only; F3 never contacts endpoints, even when false")
	if err := flags.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "doctor takes no positional arguments")
		return 2
	}
	r := doctor.Inspect()
	if *jsonOutput {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(r); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	} else {
		if _, err := fmt.Fprintf(out, "Familiar %s readiness: %s (schema %d)\n", r.Scope, r.Status, r.SchemaVersion); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		for _, check := range r.Checks {
			if _, err := fmt.Fprintf(out, "%s: %s — %s\n", check.ID, check.Status, check.Reason); err != nil {
				fmt.Fprintln(stderr, err)
				return 2
			}
		}
	}
	if r.Status != doctor.Ready {
		return 1
	}
	return 0
}
