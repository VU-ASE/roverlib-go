package rover

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// A service name can only contain lowercase letters and numbers
func ValidateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name is empty")
	} else if !regexp.MustCompile(`^[a-z0-9]*$`).MatchString(name) {
		return fmt.Errorf("service name can only contain lowercase letters and numbers")
	}

	return nil
}

// A service author can only contain letters and numbers
func ValidateServiceAuthor(author string) error {
	if author == "" {
		return fmt.Errorf("service author is empty")
	} else if !regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(author) {
		return fmt.Errorf("service author can only contain letters and numbers")
	}

	return nil
}

// A service source should not include a scheme and should be a valid URL
func ValidateServiceSource(source string) error {
	return nil // todo
}

// A service version should be a valid semantic version
func ValidateServiceVersion(version string) error {
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
	case int:
		// no checks needed
	case float64:
		// no checks needed
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

func (input ServiceInput) validate() error {
	// Name not empty?
	if input.Service == "" {
		return fmt.Errorf("input name is empty")
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
	err := ValidateServiceName(service.Name)
	if err != nil {
		return err
	}
	err = ValidateServiceAuthor(service.Author)
	if err != nil {
		return err
	}
	err = ValidateServiceSource(service.Source)
	if err != nil {
		return err
	}
	err = ValidateServiceVersion(service.Version)
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
