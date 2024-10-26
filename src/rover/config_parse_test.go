package rover

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const roverTestfilesBase = testfilesLocation + "roveryaml"

//
// Basic tests for the validate functions
//

func TestValidateEnabledService(t *testing.T) {
	// Valid inputs
	err := validateConfigEnabledService("/home/user/rover")
	assert.Nil(t, err)

	err = validateConfigEnabledService("../home/rover")
	assert.Nil(t, err)

	err = validateConfigEnabledService("home/rover")
	assert.Nil(t, err)

	err = validateConfigEnabledService("rover")
	assert.Nil(t, err)

	// Invalid inputs
	err = validateConfigEnabledService("")
	assert.NotNil(t, err)

	err = validateConfigEnabledService("")
	assert.NotNil(t, err)

	err = validateConfigEnabledService("a rover")
	assert.NotNil(t, err)
}

func TestValidateDownloadedService(t *testing.T) {
	// Valid inputs
	valid := []DownloadedService{
		{
			Name:    "service1",
			Source:  "example.com",
			Version: "1.0.0",
			Sha:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			// No sha
			Name:    "param2",
			Source:  "example.com",
			Version: "1.0.0",
		},
		{
			Name:    "param3",
			Source:  "example.com",
			Version: "1.2.3",
		},
	}
	for _, option := range valid {
		err := option.validate()
		assert.Nil(t, err, "Error validating valid downloaded service: %v", option)
	}

	// Invalid options
	invalid := []DownloadedService{
		{
			Name:    "",
			Source:  "example.com",
			Version: "1.0.0",
			Sha:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde",
		},
		{
			Name:    "param5",
			Source:  "https://web.com", // Invalid source due to scheme
			Version: "1.0.0",
			Sha:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			Name:    "param6",
			Source:  "example.com",
			Version: "1.2.3.4",
			Sha:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			Name:   "param7",
			Source: "example.com",
			// No version
			Sha: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}
	for _, option := range invalid {
		err := option.validate()
		assert.NotNil(t, err, "Error validating invalid downloaded service: %v", option)
	}
}

//
// Test for yaml validation from files and directories
//

func TestValidateRoverFilesValid(t *testing.T) {
	// Read all files and folders in the root directory
	dir := filepath.Join(roverTestfilesBase, "valid")

	// Get all files and folders in the directory
	files, err := os.ReadDir(dir)
	assert.Nil(t, err)

	// Check each file
	for _, file := range files {
		path := filepath.Join(dir, filepath.Base(file.Name()))
		// Check file exists
		_, err := os.Stat(path)
		assert.Nil(t, err, "Error checking file: %s", path)
		_, err = ParseConfigFrom(path)
		assert.Nil(t, err, "Error parsing valid file: %s", path)
	}
}

func TestValidateRoverFilesInalid(t *testing.T) {
	// Read all files and folders in the root directory
	dir := filepath.Join(roverTestfilesBase, "invalid")

	// Get all files and folders in the directory
	files, err := os.ReadDir(dir)
	assert.Nil(t, err)

	// Check each file
	for _, file := range files {
		path := filepath.Join(dir, filepath.Base(file.Name()))
		// Check file exists
		_, err := os.Stat(path)
		assert.Nil(t, err, "Error checking file: %s", path)
		_, err = ParseConfigFrom(path)
		assert.NotNil(t, err, "Error parsing invalid file: %s", path)
	}

	// Try non-existing file
	_, err = ParseConfigFrom("non-existing-file.yaml")
	assert.NotNil(t, err)
}
