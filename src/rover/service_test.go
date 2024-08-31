package rover

import (
	"fmt"
	"testing"
)

const testfilesLocation = "../../testfiles"

// Test basic validation of individual components
func TestValidateServiceName(t *testing.T) {
	err := ValidateServiceName("name")
	if err != nil {
		t.Errorf("Error validating valid name: %s", err)
	}

	err = ValidateServiceName("My first Service")
	if err == nil {
		t.Errorf("Error validating invalid name")
	}

	err = ValidateServiceName("My-first-Service")
	if err == nil {
		t.Errorf("Error validating invalid name")
	}

	err = ValidateServiceName("imaginservicev2")
	if err != nil {
		t.Errorf("Error validating valid name: %s", err)
	}
}

func TestValidateAuthor(t *testing.T) {
	err := ValidateServiceAuthor("administrator")
	if err != nil {
		t.Errorf("Error validating valid author: %s", err)
	}

	err = ValidateServiceAuthor("John Doe")
	if err != nil {
		t.Errorf("Error validating valid author: %s", err)
	}

	err = ValidateServiceAuthor("John")
	if err == nil {
		t.Errorf("Error validating invalid author")
	}

	err = ValidateServiceAuthor("John Doe Jr.")
	if err == nil {
		t.Errorf("Error validating invalid author")
	}

	err = ValidateServiceAuthor("John Doe Sr.")
	if err == nil {
		t.Errorf("Error validating invalid author")
	}
}

func TestValidateServiceSource(t *testing.T) {
	err := ValidateServiceSource("https://example.com")
	if err == nil {
		t.Errorf("Error validating invalid source")
	}

	err = ValidateServiceSource("example.com")
	if err != nil {
		t.Errorf("Error validating valid source: %s", err)
	}
}

func TestValidateServiceVersion(t *testing.T) {
	err := ValidateServiceVersion("1.0.0")
	if err != nil {
		t.Errorf("Error validating valid version: %s", err)
	}

	err = ValidateServiceVersion("1.0")
	if err == nil {
		t.Errorf("Error validating invalid version")
	}

	err = ValidateServiceVersion("1.0.0-alpha")
	if err == nil {
		t.Errorf("Error validating invalid version")
	}

	err = ValidateServiceVersion("1.0.0+build")
	if err == nil {
		t.Errorf("Error validating invalid version")
	}
}

func TestValidateOption(t *testing.T) {
	option := ServiceOption{
		Name:      "param1",
		Value:     "4",
		valueType: "int",
	}

	err := option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(int); !ok {
		t.Errorf("Error validating valid option: value not set to int")
	}

	option = ServiceOption{
		Name:      "param2",
		Value:     "teststr",
		valueType: "string",
	}

	err = option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(string); !ok {
		t.Errorf("Error validating valid option: value not set to string")
	}

	option = ServiceOption{
		Name:      "param3",
		Value:     "",
		valueType: "string",
	}

	err = option.validate()
	if err == nil {
		t.Errorf("Error validating invalid option")
	}

	option = ServiceOption{
		Name:      "param4",
		Value:     4,
		valueType: "int",
	}

	err = option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(int); !ok {
		t.Errorf("Error validating valid option: value not set to int")
	}

	option = ServiceOption{
		Name:      "param5",
		Value:     4.0,
		valueType: "float",
	}

	err = option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(float64); !ok {
		t.Errorf("Error validating valid option: value not set to float")
	}

	// Check autoparse
	option = ServiceOption{
		Name:      "param6",
		Value:     "4",
		valueType: "",
	}

	err = option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(int); !ok {
		t.Errorf("Error validating valid option: autparsed value not set to int")
	}

	option = ServiceOption{
		Name:      "param7",
		Value:     "4.0",
		valueType: "",
	}

	err = option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(float64); !ok {
		t.Errorf("Error validating valid option: autparsed value not set to float")
	}

	option = ServiceOption{
		Name:      "param8",
		Value:     "test",
		valueType: "",
	}

	err = option.validate()
	if err != nil {
		t.Errorf("Error validating valid option: %s", err)
	}
	if _, ok := option.Value.(string); !ok {
		t.Errorf("Error validating valid option: autparsed value not set to string")
	}
}

// This YAML file should be parsed correctly
func TestValidYaml(t *testing.T) {
	yamlPath := testfilesLocation + "/valid.yaml"
	service, err := ParseServiceFrom(yamlPath)
	if err != nil {
		t.Errorf("Error parsing yaml: %s", err)
	} else {
		fmt.Printf("Service definition: %+v", service)
	}
}

// This YAML file should error because it does not exist
func TestInvalidPath(t *testing.T) {
	yamlPath := testfilesLocation + "/doesnotexist.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid path")
	}
}

// This YAML file should error because it is missing some values
func TestInvalidYamlWithValuesMissing(t *testing.T) {
	yamlPath := testfilesLocation + "/invalid-missing.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlWithDuplicateNames(t *testing.T) {
	yamlPath := testfilesLocation + "/invalid-duplicate-names.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlWithDuplicateAddresses(t *testing.T) {
	yamlPath := testfilesLocation + "/invalid-duplicate-addresses.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlDuplicateOptions(t *testing.T) {
	yamlPath := testfilesLocation + "/invalid-duplicate-options.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlMissingOptionDefault(t *testing.T) {
	yamlPath := testfilesLocation + "/invalid-missing-option-default.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlOptionType(t *testing.T) {
	yamlPath := testfilesLocation + "/invalid-option-type.yaml"
	_, err := ParseServiceFrom(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

// This YAML file should be parsed correctly
func TestValidOptions(t *testing.T) {
	yamlPath := testfilesLocation + "/valid2.yaml"
	sd, err := ParseServiceFrom(yamlPath)
	if err != nil {
		t.Errorf("Error parsing yaml: %s", err)
	} else {
		fmt.Printf("Service definition: %+v", sd)
	}
}
