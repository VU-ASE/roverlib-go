package rover

import (
	"fmt"
	"regexp"
	"strings"
)

// Shared parsing and validation functions
// A service source should not include a scheme and should be a valid URL
func validateSource(source string) error {
	if source == "" {
		return fmt.Errorf("source is empty")
	} else if strings.Contains(source, "://") {
		return fmt.Errorf("source should not include a scheme")
	}

	pattern := `^([a-zA-Z0-9\-_]+(\.[a-zA-Z0-9\-_]+)+)(\/[a-zA-Z0-9\-._~:\/?#[\]@!$&'()*+,;=%]*)?$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(source) {
		return fmt.Errorf("source is not a valid URL")
	}

	return nil
}
