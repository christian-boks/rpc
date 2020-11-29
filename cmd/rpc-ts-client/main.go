package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/apex/rpc/generators/tsclient"
	"github.com/apex/rpc/generators/tstypes"
	"github.com/apex/rpc/schema"
)

func main() {
	path := flag.String("schema", "schema.json", "Path to the schema file")
	fetchLibrary := flag.String("fetch-library", "node-fetch", "Module import for the fetch library")
	flag.Parse()

	s, err := schema.Load(*path)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = generate(os.Stdout, s, "client", *fetchLibrary)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

// generate implementation.
func generate(w io.Writer, s *schema.Schema, pkg, fetchLibrary string) error {
	out := fmt.Fprintf

	out(w, "// Do not edit, this file was generated by github.com/apex/rpc.\n\n")

	err := tstypes.Generate(w, s)
	if err != nil {
		return fmt.Errorf("generating types: %w", err)
	}

	err = tsclient.Generate(w, s, fetchLibrary)
	if err != nil {
		return fmt.Errorf("generating client: %w", err)
	}

	return nil
}
