package rover

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// A service name can only contain lowercase letters and numbers
func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name is empty")
	}
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(name) {
		return fmt.Errorf("service name can only contain lowercase letters and numbers and hyphens and must start and end with a letter")
	}

	return nil
}

// A service author can only contain letters and numbers
func validateServiceAuthor(author string) error {
	if author == "" {
		return fmt.Errorf("service author is empty")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`).MatchString(author) {
		return fmt.Errorf("service author can only contain letters and numbers and hyphens and must start and end with a letter")
	}

	return nil
}

// A service source should not include a scheme and should be a valid URL
func validateServiceSource(source string) error {
	return validateSource(source)
}

// A service version should be a valid semantic version
func validateServiceVersion(version string) error {
	_, err := semver.NewVersion(version)
	return err
}

func (option ServiceOption) validate() error {
	// name not empty?
	if option.Name == "" {
		return fmt.Errorf("option name is empty")
	}

	// type is one of string, int, float, check if a value was set
	switch option.Value.(type) {
	case string:
		if option.Value == "" {
			return fmt.Errorf("option '%s' has type string, but no value was set", option.Name)
		}
		if option.valueType != "string" {
			return fmt.Errorf("option '%s' has type string, but was set to '%s'", option.Name, option.valueType)
		}
	case int:
		if option.valueType != "int" {
			return fmt.Errorf("option '%s' has type int, but was set to '%s'", option.Name, option.valueType)
		}
	case float64:
		if option.valueType != "float" {
			return fmt.Errorf("option '%s' has type float, but was set to '%s'", option.Name, option.valueType)
		}
	default:
		return fmt.Errorf("option '%s' has an invalid type: %T", option.Name, option.Value)
	}

	return nil
}

func (service Service) validateOptions() error {
	// check that all options are valid
	for i, option := range service.Options {
		err := option.validate()
		if err != nil {
			return err
		}

		// does an option with the same name exist?
		for j, otherOption := range service.Options {
			if i != j && option.Name == otherOption.Name {
				return fmt.Errorf("duplicate option name: %s", option.Name)
			}
		}
	}

	return nil
}

func (c ServiceCommands) validate() error {
	// Run command must not be empty
	if c.Run == "" {
		return fmt.Errorf("run command is empty, do not know how to execute this service")
	}

	return nil
}

func (input ServiceInput) validate() error {
	// Name not empty?
	err := validateServiceName(input.Service)
	if err != nil {
		return err
	}

	if input.Author != "" {
		return validateServiceAuthor(input.Author)
	}

	// All streams unique?
	for i, stream := range input.Streams {
		for j, otherStream := range input.Streams {
			if i != j && stream == otherStream {
				return fmt.Errorf("duplicate input stream '%s' for service '%s'", stream, input.Service)
			}
		}
	}

	return nil
}

func (service Service) validateInputs() error {
	// Check that all inputs are valid
	for i, input := range service.Inputs {
		err := input.validate()
		if err != nil {
			return err
		}

		// Check if names are unique
		for j, otherInput := range service.Inputs {
			if i != j && input.Service == otherInput.Service {
				return fmt.Errorf("duplicate input service: %s", input.Service)
			}
		}
	}

	return nil
}

func (service Service) validateOutputs() error {
	// Check if names are unique
	for i, output := range service.Outputs {
		for j, otherOutput := range service.Outputs {
			if i != j && output == otherOutput {
				return fmt.Errorf("duplicate stream output name: %s", output)
			}
		}
	}

	return nil
}

// Check if a parsed service definition is valid and can be used for service discovery
func (service Service) validate() error {
	err := validateServiceName(service.Name)
	if err != nil {
		return err
	}
	err = validateServiceAuthor(service.Author)
	if err != nil {
		return err
	}
	err = validateServiceSource(service.Source)
	if err != nil {
		return err
	}
	err = validateServiceVersion(service.Version)
	if err != nil {
		return err
	}

	err = service.Commands.validate()
	if err != nil {
		return err
	}

	err = service.validateInputs()
	if err != nil {
		return err
	}

	err = service.validateOutputs()
	if err != nil {
		return err
	}

	err = service.validateOptions()
	return err
}
