package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	terraform_plugin_schemagen "github.com/k-yomo/terraform-plugin-schemagen"
)

var version string

func main() {
	if err := runCmd(); err != nil {
		color.Red("%+v", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func runCmd() error {
	if err := os.MkdirAll("schemagen", 0755); err != nil {
		return fmt.Errorf("create output 'schema' directory: %w", err)
	}
	outputPath := filepath.Join("schemagen", "schema_gen.go")
	schema, err := terraform_plugin_schemagen.TerraformProviderSchema(context.Background())
	if err != nil {
		return fmt.Errorf("get provider schema: %w", err)
	}
	out, err := terraform_plugin_schemagen.Generate(schema)
	if err != nil {
		return fmt.Errorf("generate schema helper code: %w", err)
	}

	writer, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer writer.Close()

	if _, err := writer.Write(out); err != nil {
		return fmt.Errorf("write schema helper code: %w", err)
	}
	return nil
}
