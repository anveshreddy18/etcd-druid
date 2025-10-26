package printer

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

type Formatter interface {
	Format(data interface{}) ([]byte, error)
}

type JSONFormatter struct {
	Indent bool
}

func (f *JSONFormatter) Format(data interface{}) ([]byte, error) {
	if f.Indent {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(data interface{}) ([]byte, error) {
	return yaml.Marshal(data)
}

type TableFormatter struct{}

func (f *TableFormatter) Format(data interface{}) ([]byte, error) {
	// Implement table formatting logic here
	return nil, nil
}

func NewFormatter(format string) (Formatter, error) {
	switch format {
	case "json":
		return &JSONFormatter{Indent: true}, nil
	case "json-raw":
		return &JSONFormatter{Indent: false}, nil
	case "yaml":
		return &YAMLFormatter{}, nil
	case "":
		return nil, nil // No formatting
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}
