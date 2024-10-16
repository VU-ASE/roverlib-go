package rover

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

func validateConfigEnabledService(path string) error {
	// Service name can only be alphanumeric, hyphens and underscores
	if path == "" {
		return fmt.Errorf("enabled path is empty")
	}

	// Can only contain alphanumeric characters, hyphens, underscores, dots and slashes
	if !regexp.MustCompile(`^[a-zA-Z0-9\-_\.\/]+$`).MatchString(path) {
		return fmt.Errorf("enabled path can only contain alphanumeric characters, hyphens, underscores, dots and slashes")
	}

	return nil
}

func (s DownloadedService) validate() error {
	// Service name can only be alphanumeric, hyphens and underscores
	if s.Name == "" {
		return fmt.Errorf("service name is empty")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`).MatchString(s.Name) {
		return fmt.Errorf("service name can only contain alphanumeric characters, hyphens and underscores")
	}

	// Service source should not include a scheme and should be a valid URL
	if err := validateSource(s.Source); err != nil {
		return err
	}

	// Service version should be a valid semantic version
	if _, err := semver.NewVersion(s.Version); err != nil {
		return fmt.Errorf("service version is not a valid semantic version")
	}

	// Service sha should be a valid sha256 hash if set
	if s.Sha != "" {
		if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(s.Sha) {
			return fmt.Errorf("service sha is not a valid sha256 hash")
		}
	}

	return nil
}

func (c Config) Validate() error {
	for _, s := range c.Downloaded {
		if err := s.validate(); err != nil {
			return err
		}
	}

	for _, path := range c.Enabled {
		if err := validateConfigEnabledService(path); err != nil {
			return err
		}
	}

	// Check all enabled paths are unique
	enabled := make(map[string]struct{})
	for _, path := range c.Enabled {
		if _, ok := enabled[path]; ok {
			return fmt.Errorf("enabled path '%s' is not unique", path)
		}
		enabled[path] = struct{}{}
	}

	return nil
}
