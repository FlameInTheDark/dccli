package dprint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	// FormatTable outputs data as a formatted table (default)
	FormatTable OutputFormat = "table"
	// FormatJSON outputs data as JSON
	FormatJSON OutputFormat = "json"
	// FormatYAML outputs data as YAML
	FormatYAML OutputFormat = "yaml"
)

// OutputManager handles output formatting for different data types
type OutputManager struct {
	format OutputFormat
	writer io.Writer
}

// OutputOption configures the OutputManager
type OutputOption func(*OutputManager)

// WithFormat sets the output format
func WithFormat(format OutputFormat) OutputOption {
	return func(o *OutputManager) {
		o.format = format
	}
}

// WithWriter sets the output writer (defaults to os.Stdout)
func WithWriter(w io.Writer) OutputOption {
	return func(o *OutputManager) {
		o.writer = w
	}
}

// NewOutputManager creates a new OutputManager with the given options
func NewOutputManager(opts ...OutputOption) *OutputManager {
	om := &OutputManager{
		format: FormatTable,
		writer: os.Stdout,
	}
	for _, opt := range opts {
		opt(om)
	}
	return om
}

// SetFormat changes the output format
func (o *OutputManager) SetFormat(format OutputFormat) {
	o.format = format
}

// GetFormat returns the current output format
func (o *OutputManager) GetFormat() OutputFormat {
	return o.format
}

// Print outputs data based on the configured format
// For table format, use PrintTable with headers and rows
// For JSON/YAML, pass a struct, slice, or map
func (o *OutputManager) Print(data interface{}) error {
	switch o.format {
	case FormatJSON:
		return o.PrintJSON(data)
	case FormatYAML:
		return o.PrintYAML(data)
	default:
		return fmt.Errorf("use PrintTable for table format")
	}
}

// PrintJSON outputs data as formatted JSON
func (o *OutputManager) PrintJSON(data interface{}) error {
	encoder := json.NewEncoder(o.writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// PrintYAML outputs data as formatted YAML
func (o *OutputManager) PrintYAML(data interface{}) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	if _, err := o.writer.Write(bytes); err != nil {
		return fmt.Errorf("failed to write YAML: %w", err)
	}
	return nil
}

// PrintTable outputs data as a formatted table
func (o *OutputManager) PrintTable(headers []string, rows [][]string) error {
	if o.format != FormatTable {
		return fmt.Errorf("PrintTable can only be used with table format")
	}

	// Use the existing Table function but redirect output
	// We need to capture the output since Table prints directly to stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Table(headers, rows)

	w.Close()
	os.Stdout = oldStdout

	// Copy the captured output to our writer
	if _, err := io.Copy(o.writer, r); err != nil {
		return fmt.Errorf("failed to write table: %w", err)
	}

	return nil
}

// PrintTableFromStructs outputs a slice of structs as a table
// Uses struct tags to determine column names
func (o *OutputManager) PrintTableFromStructs(data interface{}) error {
	// This is a placeholder for future implementation
	// Would use reflection to convert structs to table rows
	return fmt.Errorf("PrintTableFromStructs not yet implemented")
}

// DefaultOutputManager is the global default output manager
var DefaultOutputManager = NewOutputManager()

// SetGlobalFormat sets the format for the default output manager
func SetGlobalFormat(format OutputFormat) {
	DefaultOutputManager.SetFormat(format)
}

// Print uses the default output manager to print data
func Print(data interface{}) error {
	return DefaultOutputManager.Print(data)
}

// PrintJSON uses the default output manager to print JSON
func PrintJSON(data interface{}) error {
	return DefaultOutputManager.PrintJSON(data)
}

// PrintYAML uses the default output manager to print YAML
func PrintYAML(data interface{}) error {
	return DefaultOutputManager.PrintYAML(data)
}

// ParseFormat parses a string into an OutputFormat
// Falls back to table format for unknown formats
func ParseFormat(s string) (OutputFormat, error) {
	switch s {
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	default:
		// Return table as default for invalid formats
		return FormatTable, fmt.Errorf("invalid output format: %s (valid: table, json, yaml)", s)
	}
}

// String returns the string representation of the format
func (f OutputFormat) String() string {
	return string(f)
}
