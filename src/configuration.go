// Structs and types for configuration of the service
// A single configuration object is exposed to the user program for transparently fetching configuration values.

package roverlib

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ServiceConfiguration struct {
	// Managed per type, because Go does not support easy union types
	intOptions    map[string]int
	floatOptions  map[string]float64
	stringOptions map[string]string
	tunable       map[string]bool
	// For concurrency control
	lock *sync.RWMutex
	// Prevent late updates
	lastUpdate uint64 // timestamp
}

func NewServiceConfiguration(service Service) *ServiceConfiguration {
	config := &ServiceConfiguration{
		intOptions:    make(map[string]int),
		floatOptions:  make(map[string]float64),
		stringOptions: make(map[string]string),
		tunable:       make(map[string]bool),
		lastUpdate:    uint64(time.Now().UnixMilli()),
	}

	for _, c := range service.Configuration {
		switch *c.Type {
		case Int:
			config.intOptions[*c.Name] = int(*c.Value.Integer)
		case Float:
			config.floatOptions[*c.Name] = *c.Value.Double
		case String:
			config.stringOptions[*c.Name] = *c.Value.String
		}
		if c.Tunable != nil {
			config.tunable[*c.Name] = *c.Tunable
		}
	}

	return config
}

//
// Methods accessible by the user program
// nb: we force the user to be very explicit about the type of the configuration value they want to fetch, to avoid runtime errors
//

// Returns the integer value of the configuration option with the given name, returns an error if the option does not exist or does not exist for this type
// Reading is NOT thread-safe, but we accept the risks because we assume that the user program will read the configuration values repeatedly
// If you want to read the configuration values concurrently, you should use the GetIntSafe method
func (c *ServiceConfiguration) GetInt(name string) (int, error) {
	value, ok := c.intOptions[name]
	if !ok {
		return 0, fmt.Errorf("no integer configuration option with name %s", name)
	}
	return value, nil
}

func (c *ServiceConfiguration) GetIntSafe(name string) (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.GetInt(name)
}

// Returns the float value of the configuration option with the given name, returns an error if the option does not exist or does not exist for this type
// Reading is NOT thread-safe, but we accept the risks because we assume that the user program will read the configuration values repeatedly
// If you want to read the configuration values concurrently, you should use the GetFloatSafe method
func (c *ServiceConfiguration) GetFloat(name string) (float64, error) {
	value, ok := c.floatOptions[name]
	if !ok {
		return 0, fmt.Errorf("no float configuration option with name %s", name)
	}
	return value, nil
}

func (c *ServiceConfiguration) GetFloatSafe(name string) (float64, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.GetFloat(name)
}

// Returns the string value of the configuration option with the given name, returns an error if the option does not exist or does not exist for this type
// Reading is NOT thread-safe, but we accept the risks because we assume that the user program will read the configuration values repeatedly
// If you want to read the configuration values concurrently, you should use the GetStringSafe method
func (c *ServiceConfiguration) GetString(name string) (string, error) {
	value, ok := c.stringOptions[name]
	if !ok {
		return "", fmt.Errorf("no string configuration option with name %s", name)
	}
	return value, nil
}

func (c *ServiceConfiguration) GetStringSafe(name string) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.GetString(name)
}

//
// Methods for internal use
//

// Set the integer value of the configuration option with the given name (thread-safe)
func (c *ServiceConfiguration) setInt(name string, value int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.tunable[name] {
		c.intOptions[name] = value
		log.Debug().Str("name", name).Int("value", value).Msg("Set integer configuration option")
	} else {
		log.Debug().Str("name", name).Msg("Attempted to set non-tunable integer configuration option")
	}
}

// Set the float value of the configuration option with the given name (thread-safe)
func (c *ServiceConfiguration) setFloat(name string, value float64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.tunable[name] {
		c.floatOptions[name] = value
		log.Debug().Str("name", name).Float64("value", value).Msg("Set float configuration option")
	} else {
		log.Debug().Str("name", name).Msg("Attempted to set non-tunable float configuration option")
	}
}

// Set the string value of the configuration option with the given name (thread-safe)
func (c *ServiceConfiguration) setString(name string, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.tunable[name] {
		c.stringOptions[name] = value
		log.Debug().Str("name", name).Str("value", value).Msg("Set string configuration option")
	} else {
		log.Debug().Str("name", name).Msg("Attempted to set non-tunable string configuration option")
	}
}
