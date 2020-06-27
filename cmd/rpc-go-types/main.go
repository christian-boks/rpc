package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/apex/rpc/generators/gotypes"
	"github.com/apex/rpc/schema"
)

func main() {
	path := flag.String("schema", "schema.json", "Path to the schema file")
	pkg := flag.String("package", "api", "Name of the package")
	flag.Parse()

	s, err := schema.Load(*path)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = generate(os.Stdout, s, *pkg)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

// generate implementation.
func generate(w io.Writer, s *schema.Schema, pkg string) error {
	out := fmt.Fprintf

	// TODO: move these to generator
	out(w, "// Do not edit, this file was generated by github.com/apex/rpc.\n\n")
	out(w, "package %s\n\n", pkg)

	out(w, "import (\n")
	out(w, "  \"fmt\"\n")
	out(w, "  \"time\"\n")
	out(w, "\n")
	out(w, "  \"github.com/apex/rpc\"\n")
	out(w, ")\n\n")

	err := gotypes.Generate(w, s, true)
	if err != nil {
		return fmt.Errorf("generating types: %w", err)
	}

	return nil
}
