package servicerunner

import (
	"fmt"
	"testing"
)

// This YAML file should be parsed correctly
func TestValidYaml(t *testing.T) {
	yamlPath := "../testfiles/valid.yaml"
	serviceDefinition, err := parseServiceDefinitionFromYaml(yamlPath)
	if err != nil {
		t.Errorf("Error parsing yaml: %s", err)
	} else {
		fmt.Printf("Service definition: %+v", serviceDefinition)
	}
}

// This YAML file should error
func TestInvalidPath(t *testing.T) {
	yamlPath := "../testfiles/doesnotexist.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid path")
	}
}

func TestInvalidYamlWithValuesMissing(t *testing.T) {
	yamlPath := "../testfiles/invalid-missing.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlWithDuplicateNames(t *testing.T) {
	yamlPath := "../testfiles/invalid-duplicate-names.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlWithDuplicateAddresses(t *testing.T) {
	yamlPath := "../testfiles/invalid-duplicate-addresses.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlDuplicateOptions(t *testing.T) {
	yamlPath := "../testfiles/invalid-duplicate-options.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlMissingOptionDefault(t *testing.T) {
	yamlPath := "../testfiles/invalid-missing-option-default.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestInvalidYamlOptionType(t *testing.T) {
	yamlPath := "../testfiles/invalid-option-type.yaml"
	_, err := parseServiceDefinitionFromYaml(yamlPath)
	if err == nil {
		t.Errorf("Parsed invalid yaml")
	}
}

func TestValidOptions(t *testing.T) {
	yamlPath := "../testfiles/valid2.yaml"
	sd, err := parseServiceDefinitionFromYaml(yamlPath)
	if err != nil {
		t.Errorf("Error parsing yaml: %s", err)
	} else {
		fmt.Printf("Service definition: %+v", sd)
	}
}
