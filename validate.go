// Command lt-validate checks that each non-blank line of one or more .jsonl
// files is a compliant LudoTrace event. It is the executable form of SPEC.md's
// Line Format and Reserved Event Types sections.
//
// The contract (event.schema.json) is embedded, so the built binary is fully
// self-contained: download one file, run it, no toolchain or dependencies.
//
// Usage:
//
//	lt-validate <file.jsonl> [more.jsonl ...]
//
// Exit code 0 = all lines valid, 1 = one or more violations (or bad args).
package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var printer = message.NewPrinter(language.English)

//go:embed event.schema.json
var schemaJSON []byte

// version is stamped at build time via -ldflags "-X main.version=<VERSION>".
// It defaults to "dev" for local `go build` / `go run`.
var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("lt-validate %s\n", version)
		return
	}

	files := args
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "usage: lt-validate <file.jsonl> [more.jsonl ...]")
		fmt.Fprintln(os.Stderr, "       lt-validate --version")
		os.Exit(1)
	}

	schema := mustCompileSchema()

	totalViolations := 0
	for _, file := range files {
		totalViolations += validateFile(schema, file)
	}
	if totalViolations != 0 {
		os.Exit(1)
	}
}

func mustCompileSchema() *jsonschema.Schema {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal: embedded schema is not valid JSON: %v\n", err)
		os.Exit(1)
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat() // treat date / date-time format as a constraint, not just an annotation
	if err := c.AddResource("event.schema.json", doc); err != nil {
		fmt.Fprintf(os.Stderr, "internal: %v\n", err)
		os.Exit(1)
	}
	schema, err := c.Compile("event.schema.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal: cannot compile embedded schema: %v\n", err)
		os.Exit(1)
	}
	return schema
}

func validateFile(schema *jsonschema.Schema, file string) int {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
		return 1
	}
	defer f.Close()

	violations := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // allow long lines (large snapshots)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			fmt.Fprintf(os.Stderr, "%s:%d: blank line (SPEC: no blank lines)\n", file, lineNo)
			violations++
			continue
		}

		inst, err := jsonschema.UnmarshalJSON(strings.NewReader(line))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s:%d: not valid JSON — %v\n", file, lineNo, err)
			violations++
			continue
		}

		if err := schema.Validate(inst); err != nil {
			for _, msg := range leafErrors(err) {
				fmt.Fprintf(os.Stderr, "%s:%d: %s\n", file, lineNo, msg)
			}
			violations++
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: read error: %v\n", file, err)
		violations++
	}

	if violations == 0 {
		fmt.Printf("%s: OK\n", file)
	} else {
		fmt.Printf("%s: %d invalid line(s)\n", file, violations)
	}
	return violations
}

// leafErrors flattens a ValidationError tree to the actionable leaf causes,
// dropping the "doesn't validate with .../allOf/N/then" wrapper nodes that the
// if/then reserved-type branches produce.
func leafErrors(err error) []string {
	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return []string{err.Error()}
	}
	var out []string
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			loc := strings.Join(e.InstanceLocation, "/")
			if loc == "" {
				loc = "(root)"
			} else {
				loc = "/" + loc
			}
			out = append(out, fmt.Sprintf("%s %s", loc, e.ErrorKind.LocalizedString(printer)))
			return
		}
		for _, c := range e.Causes {
			walk(c)
		}
	}
	walk(ve)
	return out
}
