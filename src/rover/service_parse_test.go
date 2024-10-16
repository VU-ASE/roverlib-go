package rover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const serviceTestfilesBase = testfilesLocation + "serviceyaml"

//
// Basic tests for the validate functions
//

func TestValidateServiceName(t *testing.T) {
	// Valid inputs
	err := validateServiceName("name")
	assert.Nil(t, err)

	err = validateServiceName("name1")
	assert.Nil(t, err)

	err = validateServiceName("name-1")
	assert.Nil(t, err)

	err = validateServiceName("my-first-service-from-vu-ase")
	assert.Nil(t, err)

	err = validateServiceName("my-first-service-from-vu-ase-1")
	assert.Nil(t, err)

	// Invalid inputs
	err = validateServiceName("name_1")
	assert.NotNil(t, err)

	err = validateServiceName("Name")
	assert.NotNil(t, err)

	err = validateServiceName("name-")
	assert.NotNil(t, err)

	err = validateServiceName("-name")
	assert.NotNil(t, err)

	err = validateServiceName("name-1-")
	assert.NotNil(t, err)

	err = validateServiceName("MYFIRSTSERVICE")
	assert.NotNil(t, err)

	err = validateServiceName("my-first-service-from-vu-ase-")
	assert.NotNil(t, err)

	err = validateServiceName("")
	assert.NotNil(t, err)
}

func TestValidateAuthor(t *testing.T) {
	// Valid inputs
	err := validateServiceAuthor("administrator")
	assert.Nil(t, err)

	err = validateServiceAuthor("JohnDoe")
	assert.Nil(t, err)

	err = validateServiceAuthor("administrat0r")
	assert.Nil(t, err)

	err = validateServiceAuthor("0JohnD0")
	assert.Nil(t, err)

	err = validateServiceAuthor("JohnDoeJr")
	assert.Nil(t, err)

	err = validateServiceAuthor("john-jr")
	assert.Nil(t, err)

	// Invalid inputs
	err = validateServiceAuthor("")
	assert.NotNil(t, err)

	err = validateServiceAuthor("John Doe")
	assert.NotNil(t, err)

	err = validateServiceAuthor("johndoe-")
	assert.NotNil(t, err)

	err = validateServiceAuthor("-johnd")
	assert.NotNil(t, err)
}

func TestValidateServiceSource(t *testing.T) {
	// Valid inputs
	err := validateServiceSource("example.com")
	assert.Nil(t, err)

	err = validateServiceSource("example.com/path")
	assert.Nil(t, err)

	err = validateServiceSource("example.com/path/to/service")
	assert.Nil(t, err)

	err = validateServiceSource("a.b.c.example.com/path/to/service/")
	assert.Nil(t, err)

	err = validateServiceSource("a.b.c.example.com/path/to/service/?query=1")
	assert.Nil(t, err)

	// Invalid inputs
	err = validateServiceSource("")
	assert.NotNil(t, err)

	err = validateServiceSource("http://example.com")
	assert.NotNil(t, err)

	err = validateServiceSource("https://example.com")
	assert.NotNil(t, err)

	err = validateServiceSource("example")
	assert.NotNil(t, err)

	err = validateServiceSource("example.")
	assert.NotNil(t, err)

	err = validateServiceSource("example.com.")
	assert.NotNil(t, err)

	err = validateServiceSource(".example.com:8080/")
	assert.NotNil(t, err)
}

func TestValidateOption(t *testing.T) {
	// Valid inputs
	valid := []ServiceOption{
		{
			Name:      "param1",
			Value:     4,
			valueType: "int",
		},
		{
			Name:      "param2",
			Value:     "teststr",
			valueType: "string",
		},
		{
			Name:      "param3",
			Value:     4.5,
			valueType: "float",
		},
	}
	for _, option := range valid {
		err := option.validate()
		assert.Nil(t, err, "Error validating valid option: %v", option)
	}

	// Invalid options
	invalid := []ServiceOption{
		{
			Name:      "param4",
			Value:     "",
			valueType: "string",
		},
		{
			Name:      "param5",
			Value:     4,
			valueType: "string",
		},
		{
			Name:      "param6",
			Value:     "4",
			valueType: "float",
		},
		{
			Name:      "",
			Value:     "",
			valueType: "string",
		},
		{
			Name:      "",
			Value:     "abc",
			valueType: "string",
		},
		{
			Name:      "param7",
			Value:     "abc",
			valueType: "nonexistenttype",
		},
		{
			Name:      "param9",
			Value:     "abc",
			valueType: "",
		},
	}
	for _, option := range invalid {
		err := option.validate()
		assert.NotNil(t, err, "Error validating invalid option: %v", option)
	}
}

//
// Test for yaml validation from files and directories
//

func TestValidateServiceFilesValid(t *testing.T) {
	// Read all files and folders in the root directory
	dir := filepath.Join(serviceTestfilesBase, "valid")

	// Get all files and folders in the directory
	files, err := os.ReadDir(dir)
	assert.Nil(t, err)

	// Check each file
	for _, file := range files {
		path := filepath.Join(dir, filepath.Base(file.Name()))
		// Check file exists
		_, err := os.Stat(path)
		assert.Nil(t, err, "Error checking file: %s", path)
		_, err = ParseServiceFrom(path)
		assert.Nil(t, err, "Error parsing valid file: %s", path)
	}
}

func TestValidateServiceFilesInvalid(t *testing.T) {
	// Read all files and folders in the root directory
	dir := filepath.Join(serviceTestfilesBase, "invalid")

	// Get all files and folders in the directory
	files, err := os.ReadDir(dir)
	assert.Nil(t, err)

	// Check each file
	for _, file := range files {
		path := filepath.Join(dir, filepath.Base(file.Name()))
		// Check file exists
		_, err := os.Stat(path)
		assert.Nil(t, err, "Error checking file: %s", path)
		_, err = ParseServiceFrom(path)
		assert.NotNil(t, err, "Error parsing invalid file: %s", path)
	}

	// Try non-existing file
	_, err = ParseServiceFrom("non-existing-file.yaml")
	assert.NotNil(t, err)
}
