package rover

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Configuration of a Rover service as defined in a service.yaml file
type Service struct {
	Name     string          `yaml:"name"`
	Author   string          `yaml:"author"`
	Source   string          `yaml:"source"`
	Version  string          `yaml:"version"`
	Inputs   []ServiceInput  `yaml:"inputs"`
	Outputs  []string        `yaml:"outputs"`
	Options  []ServiceOption `yaml:"configuration"`
	Commands ServiceCommands `yaml:"commands"`
}

type ServiceCommands struct {
	Build string `yaml:"build"` // optional
	Run   string `yaml:"run"`
}

type ServiceInput struct {
	Service string   `yaml:"service"`
	Author  string   // This is not explicitly defined in the YAML file, but users can specify a name like vu-ase/imaging, which will be split into Author: vu-ase and Service: imaging
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

	// Make sure to split the author from the name if it is specified
	for i, input := range service.Inputs {
		parts := strings.Split(input.Service, "/")
		if len(parts) == 2 {
			service.Inputs[i].Author = parts[0]
			service.Inputs[i].Service = parts[1]
		}
	}

	// Make sure to set all values to lowercase values (except configuration options)
	service.Name = strings.ToLower(service.Name)
	service.Author = strings.ToLower(service.Author)
	for i, input := range service.Inputs {
		service.Inputs[i].Service = strings.ToLower(input.Service)
		service.Inputs[i].Author = strings.ToLower(input.Author)
		for j, stream := range input.Streams {
			service.Inputs[i].Streams[j] = strings.ToLower(stream)
		}
	}
	for i, output := range service.Outputs {
		service.Outputs[i] = strings.ToLower(output)
	}
	for i, option := range service.Options {
		service.Options[i].Name = strings.ToLower(option.Name)
	}

	err = service.validate()
	return service, err
}

// Parse a service.yaml from a file path. This path can either be a yaml file or a directory containing a service.yaml file
// if a directory contains multiple service.yaml files, one must be explicitly specified,
// It will only look in the root of the directory and not recurse into subdirectories
func ParseServiceFrom(path string) (*Service, error) {
	yamlPath := path

	// Is the path a directory?
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		// Check if the directory contains files matching the service*.yaml pattern
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		matches := []string{}

		// Pattern explanation: `(?i)` makes the pattern case insensitive,
		// `^service.*\.yaml$` matches strings starting with "service",
		// followed by any number of characters, and ending with ".yaml".
		pattern := `(?i)^service.*\.yaml$`
		regexp := regexp.MustCompile(pattern)

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			matched := regexp.MatchString(file.Name())
			if matched {
				matches = append(matches, file.Name())
			}
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("no service.yaml file found in directory: %s. Specify an explicit path to a service.yaml or make sure to include a service.yaml in this directory.", path)
		}
		if len(matches) > 1 {
			str := matches[0]
			for _, match := range matches[1:] {
				str += ", " + match
			}

			return nil, fmt.Errorf("multiple service.yaml files found in directory: %s. Specify an explicit path to a service.yaml or remove the service.yaml files that you don't use. Found: %s", path, str)
		}

		yamlPath = filepath.Join(path, filepath.Base(matches[0]))
	}

	// Read the file
	content, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	return ParseService(content)
}

func (opt *ServiceOption) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Temporary structure to unmarshal the type and raw value
	var temp struct {
		Name  string      `yaml:"name"`
		Type  string      `yaml:"type"`
		Value interface{} `yaml:"value"`
	}

	// Unmarshal into the temporary structure
	if err := unmarshal(&temp); err != nil {
		return err
	}

	opt.Name = temp.Name
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
	case "":
		// Autodetect the type
		opt.Value = temp.Value
		switch temp.Value.(type) {
		case string:
			temp.Type = "string"
		case int:
			temp.Type = "int"
		case float64:
			temp.Type = "float"
		default:
			return fmt.Errorf("unsupported autoparse type: %T", temp.Value)
		}
	default:
		return fmt.Errorf("unsupported explicit type: %s", temp.Type)
	}
	opt.valueType = temp.Type

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
