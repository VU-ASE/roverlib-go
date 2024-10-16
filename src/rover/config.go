package rover

import (
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

// Configuration of a Rover, as defined in a rover.yaml file
type Config struct {
	Downloaded []DownloadedService `yaml:"downloaded"`
	// Every string is a path to a service on the Rover. The path should lead to a service.yaml file
	Enabled []string `yaml:"enabled"`
}

// Describes a downloaded service from a given source
type DownloadedService struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Version string `yaml:"version"`
	Sha     string `yaml:"sha"` // optional
}

func (c *Config) Enable(path string) {
	if c == nil {
		return
	}

	c.Enabled = append(c.Enabled, path)
}

func (c *Config) Disable(path string) {
	if c == nil {
		return
	}

	c.Enabled = slices.DeleteFunc(
		c.Enabled,
		func(p string) bool {
			return p == path
		},
	)
}

func (c *Config) Toggle(path string) {
	if c.HasEnabled(path) {
		c.Disable(path)
	} else {
		c.Enable(path)
	}
}

func (c *Config) HasEnabled(path string) bool {
	if c == nil {
		return false
	}

	for _, p := range c.Enabled {
		if p == path {
			return true
		}
	}

	return false
}

// Parse rover.yaml contents from a byte array
// NB: having a custom Parse function allows us to set default values for the struct
func ParseConfig(content []byte) (*Config, error) {
	config := &Config{}
	err := yaml.Unmarshal(content, config)
	if err != nil {
		return nil, err
	}

	// Make sure all paths are minimal
	for i, path := range config.Enabled {
		config.Enabled[i] = filepath.Clean(path)
	}

	err = config.Validate()
	return config, err
}

// Parse a rover.yaml from a file path
func ParseConfigFrom(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseConfig(content)
}
