package rover

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Configuration of a Rover service as defined in a service.yaml file
type Service struct {
	Name    string          `yaml:"name"`
	Author  string          `yaml:"author"`
	Source  string          `yaml:"source"`
	Version string          `yaml:"version"`
	Inputs  []ServiceInput  `yaml:"inputs"`
	Outputs []string        `yaml:"outputs"`
	Options []ServiceOption `yaml:"configuration"`
}

type ServiceInput struct {
	Service string   `yaml:"service"`
	Streams []string `yaml:"streams"`
}

type ServiceOption struct {
	Name      string      `yaml:"name"`
	Value     interface{} `yaml:"value"`
	valueType string      `yaml:"type"` // string, int, float or undefined for autoparsing
	Tunable   bool        `yaml:"tunable"`
}

// Parse service.yaml contents from a byte array
// NB: having a custom Parse function allows us to set default values for the struct
func ParseService(content []byte) (*Service, error) {
	service := &Service{}
	err := yaml.Unmarshal(content, service)
	if err != nil {
		return nil, err
	}
	err = service.validate()
	return service, err
}

// Parse a service.yaml from a file path
func ParseServiceFrom(path string) (*Service, error) {
	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseService(content)
}

func (opt *ServiceOption) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Temporary structure to unmarshal the type and raw value
	var temp struct {
		Type  string      `yaml:"type"`
		Value interface{} `yaml:"value"`
	}

	// Unmarshal into the temporary structure
	if err := unmarshal(&temp); err != nil {
		return err
	}

	opt.valueType = temp.Type
	// Dynamically unmarshal based on the type
	switch temp.Type {
	case "string":
		var s string
		if err := unmarshalField(temp.Value, &s); err != nil {
			return err
		}
		opt.Value = s
	case "int":
		var i int
		if err := unmarshalField(temp.Value, &i); err != nil {
			return err
		}
		opt.Value = i
	case "float":
		var f float64
		if err := unmarshalField(temp.Value, &f); err != nil {
			return err
		}
		opt.Value = f
	default:
		return fmt.Errorf("unsupported type: %s", temp.Type)
	}

	return nil
}

// Helper function to unmarshal individual fields
func unmarshalField(input interface{}, out interface{}) error {
	bytes, err := yaml.Marshal(input)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, out)
}
