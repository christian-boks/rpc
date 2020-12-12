package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/apex/rpc/generators/rustclient"
	"github.com/apex/rpc/generators/rusttypes"
	"github.com/apex/rpc/schema"
)

//
// The following dependencies are required to use the client
//
// [dependencies]
// serde = { version = "1.0", features = ["derive"] }
// serde_json = "1.0"
// serde_derive = "1.0"
// bytes = "0.5"
// chrono = { version = "0.4", features = ["serde"] }
// reqwest = { version = "0.10", features = ["json"] }
// tokio = { version = "0.2", features = ["full"] }
//

func main() {
	path := flag.String("schema", "schema.json", "Path to the schema file")
	flag.Parse()

	s, err := schema.Load(*path)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = generate(os.Stdout, s)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

// generate implementation.
func generate(w io.Writer, s *schema.Schema) error {
	out := fmt.Fprintf

	// force tags to be json only
	s.Go.Tags = []string{"json"}

	out(w, "// Do not edit, this file was generated by github.com/apex/rpc.\n\n")
	out(w, "use serde::{Deserialize, Serialize};\n")
	out(w, "use chrono::{DateTime};\n\n")

	err := rusttypes.Generate(w, s, false)
	if err != nil {
		return fmt.Errorf("generating types: %w", err)
	}

	err = rustclient.Generate(w, s)
	if err != nil {
		return fmt.Errorf("generating client: %w", err)
	}

	return nil
}
