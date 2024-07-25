package servicerunner

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-yaml/yaml"
)

//
// This file describes the structure of a service definition, as should be included in the service.yaml file.
// The service.yaml file is used to describe the service and its dependencies and needs to follow this structure.
//

type serviceDefinition struct {
	Name         string       `yaml:"name"`
	Description  string       `yaml:"description"`
	Dependencies []dependency `yaml:"dependencies"`
	Outputs      []output     `yaml:"outputs"`
	Options      []option     `yaml:"options"`
}

// A dependency definition as it should be included in the yaml file
type dependency struct {
	// name of the service that this service depends on
	ServiceName string `yaml:"service"`
	// the name of the output that this service needs from the dependency
	OutputName string `yaml:"output"`
}

// The output that this service will produce
type output struct {
	// the name of the output that this service will produce
	Name string `yaml:"name"`
	// the address of the output that this service will produce
	Address string `yaml:"address"`
}

// A configuration option for the service, that can be either set in the yaml file or be fetched from tuning
type option struct {
	// the name of the option
	Name string `yaml:"name"`
	// can this option be updated at runtime?
	Mutable bool `yaml:"mutable"`
	// the type of the option
	Type string `yaml:"type"`
	// the value of the option (if provided)
	DefaultValue string `yaml:"default"`
}

func validateServiceDefinitionOption(option option) error {
	// name not empty?
	if option.Name == "" {
		return fmt.Errorf("Option name is empty")
	}

	// correct type?
	if option.Type != "string" && option.Type != "int" && option.Type != "float" {
		return fmt.Errorf("Option '%s' type must be string, int or float (got %s)", option.Name, option.Type)
	}

	// if not mutable, default value must be set
	if !option.Mutable && option.DefaultValue == "" {
		return fmt.Errorf("Option '%s' has no default value but is also declared not mutable. Add a default value or mark this option as mutable by setting mutable: true", option.Name)
	}

	// check if default value is of the correct type
	if option.DefaultValue != "" {
		switch option.Type {
		case "string":
			// no checks needed
			break
		case "int":
			_, err := strconv.Atoi(option.DefaultValue)
			if err != nil {
				return fmt.Errorf("Option '%s' has type int, but a default value that is not an int: %s", option.Name, option.DefaultValue)
			}
		case "float":
			_, err := strconv.ParseFloat(option.DefaultValue, 64)
			if err != nil {
				return fmt.Errorf("Option '%s' has type float, but a default value that is not a float: %s", option.Name, option.DefaultValue)
			}
		}
	}

	return nil
}

func validateServiceDefinitionOptions(options []option) error {
	// check that all options are valid
	for i, option := range options {
		err := validateServiceDefinitionOption(option)
		if err != nil {
			return err
		}

		// does an option with the same name exist?
		for j, otherOption := range options {
			if i != j && option.Name == otherOption.Name {
				return fmt.Errorf("Duplicate option: %s", option.Name)
			}
		}
	}

	return nil
}

func validateServiceDefinition(serviceDefinition serviceDefinition) error {
	if serviceDefinition.Name == "" {
		return fmt.Errorf("Service name is empty")
	} else if serviceDefinition.Description == "" {
		return fmt.Errorf("Service description is empty")
	}

	if len(serviceDefinition.Dependencies) > 0 {
		for i, dependency := range serviceDefinition.Dependencies {
			if dependency.ServiceName == "" {
				return fmt.Errorf("Dependency service name is empty")
			} else if dependency.OutputName == "" {
				return fmt.Errorf("Dependency output name is empty")
			}

			// Check if service name and output name together are unique
			for j, otherDependency := range serviceDefinition.Dependencies {
				if i != j && dependency.ServiceName == otherDependency.ServiceName && dependency.OutputName == otherDependency.OutputName {
					return fmt.Errorf("Duplicate dependency: %s %s", dependency.ServiceName, dependency.OutputName)
				}
			}
		}
	}

	if len(serviceDefinition.Outputs) > 0 {
		for i, output := range serviceDefinition.Outputs {
			if output.Name == "" {
				return fmt.Errorf("Output name is empty")
			} else if output.Address == "" {
				return fmt.Errorf("Output address is empty")
			}

			// Check if names and addresses (individually) are unique
			for j, otherOutput := range serviceDefinition.Outputs {
				if i != j && output.Name == otherOutput.Name {
					return fmt.Errorf("Duplicate output name: %s", output.Name)
				}
				if i != j && output.Address == otherOutput.Address {
					return fmt.Errorf("Duplicate output address: %s", output.Address)
				}
			}
		}
	}

	err := validateServiceDefinitionOptions(serviceDefinition.Options)
	if err != nil {
		return err
	}

	return nil
}

func parseServiceDefinition(yamlString string) (serviceDefinition, error) {
	serviceDefinition := serviceDefinition{}
	err := yaml.Unmarshal([]byte(yamlString), &serviceDefinition)
	if err != nil {
		return serviceDefinition, err
	}

	validationError := validateServiceDefinition(serviceDefinition)
	if validationError != nil {
		return serviceDefinition, validationError
	}

	return serviceDefinition, validationError
}

func parseServiceDefinitionFromYaml(path string) (serviceDefinition, error) {
	// Read file
	yaml, err := os.ReadFile(path)
	if err != nil {
		return serviceDefinition{}, err
	}
	return parseServiceDefinition(string(yaml))
}
