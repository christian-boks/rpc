package rusttypes

import (
	"fmt"
	"io"
	"strings"

	"github.com/apex/rpc/internal/format"
	"github.com/apex/rpc/internal/schemautil"
	"github.com/apex/rpc/schema"
)

// Generate writes the Rust type implementations to w, with optional validation methods.
func Generate(w io.Writer, s *schema.Schema) error {
	out := fmt.Fprintf

	// default tags
	if s.Go.Tags == nil {
		s.Go.Tags = []string{"json"}
	}

	// types
	for _, t := range s.TypesSlice() {
		out(w, "// %s %s\n", format.GoName(t.Name), t.Description)
		out(w, "#[derive(Serialize, Deserialize, Debug, Clone)]\n")
		out(w, "pub struct %s {\n", format.GoName(t.Name))
		writeFields(w, s, t.Properties)
		out(w, "}\n\n")
	}

	// methods
	for _, m := range s.Methods {
		name := format.GoName(m.Name)

		// inputs
		if len(m.Inputs) > 0 {
			out(w, "// %sInput params.\n", name)
			out(w, "#[derive(Serialize, Debug, Clone)]\n")
			out(w, "pub struct %sInput{\n", name)
			writeFields(w, s, m.Inputs)
			out(w, "}\n")
		}

		// both
		if len(m.Inputs) > 0 && len(m.Outputs) > 0 {
			out(w, "\n")
		}

		// outputs
		if len(m.Outputs) > 0 {
			out(w, "// %s Output params.\n", name)
			out(w, "#[derive(Deserialize, Debug, Clone)]\n")
			out(w, "pub struct %sOutput{\n", name)
			writeFields(w, s, m.Outputs)
			out(w, "}\n")
		}

		out(w, "\n")
	}

	return nil
}

// writeFields to writer.
func writeFields(w io.Writer, s *schema.Schema, fields []schema.Field) {
	for i, f := range fields {
		writeField(w, s, f)
		if i < len(fields)-1 {
			fmt.Fprintf(w, "\n")
		}
	}
}

// writeField to writer.
func writeField(w io.Writer, s *schema.Schema, f schema.Field) {
	fmt.Fprintf(w, "  // %s is %s%s\n", format.RustName(f.Name), f.Description, schemautil.FormatExtra(f))
	if f.Required == true {
		fmt.Fprintf(w, "  pub %s: %s,\n", format.RustName(f.Name), rustType(s, f))
	} else {
		fmt.Fprintf(w, "  pub %s: Option<%s>,\n", format.RustName(f.Name), rustType(s, f))
	}
}

// rustType returns a Rust equivalent type for field f.
func rustType(s *schema.Schema, f schema.Field) string {
	// ref
	if ref := f.Type.Ref.Value; ref != "" {
		t := schemautil.ResolveRef(s, f.Type.Ref)
		return format.GoName(t.Name)
	}

	// type
	switch f.Type.Type {
	case schema.String:
		return "String"
	case schema.Int:
		return "i64"
	case schema.Bool:
		return "bool"
	case schema.Float:
		return "f64"
	case schema.Timestamp:
		return "DateTime<chrono::Utc>"
	case schema.Object:
		return "HashMap<String, std::any::Any>"
	case schema.Array:
		return "Vec<" + strings.Title(rustType(s, schema.Field{
			Type: schema.TypeObject(f.Items),
		})) + ">"
	default:
		panic("unhandled type")
	}
}

// fieldTags returns tags for a field.
func fieldTags(f schema.Field, tags []string) string {
	var pairs [][]string

	for _, tag := range tags {
		pairs = append(pairs, []string{tag, f.Name})
	}

	return formatTags(pairs)
}

// formatTags returns field tags.
func formatTags(tags [][]string) string {
	var s []string
	for _, t := range tags {
		if len(t) == 2 {
			s = append(s, fmt.Sprintf("%s:%q", t[0], t[1]))
		} else {
			s = append(s, fmt.Sprintf("%s", t[0]))
		}
	}
	return fmt.Sprintf("`%s`", strings.Join(s, " "))
}
