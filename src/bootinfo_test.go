package roverlib_test

import (
	"fmt"
	"os"
	"testing"

	roverlib "github.com/VU-ASE/roverlib-go/src"
)

// Valid service to test with
func injectValidService() {
	os.Setenv("ASE_SERVICE", `{
		"name": "controller",
		"version": "1.0.1",
		"inputs": [
			{
				"service": "imaging",
				"streams": [
					{
						"name": "track_data",
						"address": "tcp://unix:7890"
					},
					{
						"name": "debug_info",
						"address": "tcp://unix:7891"
					}
				]
			},
			{
				"service": "navigation",
				"streams": [
					{
						"name": "location_data",
						"address": "tcp://unix:7892"
					}
				]
			}
		],
		"outputs": [
			{
				"name": "motor_movement",
				"address": "tcp://unix:7882"
			},
			{
				"name": "sensor_data",
				"address": "tcp://unix:7883"
			}
		],
		"configuration": [
			{
				"name": "max-iterations",
				"type": "int",
				"tunable": true,
				"value": 100
			},
			{
				"name": "speed",
				"type": "float",
				"tunable": false,
				"value": 1.5
			},
			{
				"name": "log-level",
				"type": "string",
				"tunable": false,
				"value": "debug"
			}
		],
		"tuning": {
			"enabled": true,
			"address": "tcp://unix:8829"
		}
	}`)
}

// Test if we can start an empty program
func TestValidEmptyProgram(t *testing.T) {
	main := func(s roverlib.Service, config *roverlib.ServiceConfiguration) error {
		return nil
	}
	onTerminate := func(s os.Signal) error {
		return nil
	}
	injectValidService()

	roverlib.Run(main, onTerminate)
}

// Test if we can access the configuration values
func TestValidProgramWithConfigAccess(t *testing.T) {
	main := func(s roverlib.Service, config *roverlib.ServiceConfiguration) error {
		// Test integer access
		i, err := config.GetInt("max-iterations")
		if err != nil {
			t.Errorf("Failed to get integer configuration value: %s", err)
		}
		fmt.Printf("max-iterations: %d\n", i)
		// Test SAFE integer access
		i, err = config.GetIntSafe("max-iterations")
		if err != nil {
			t.Errorf("Failed to get integer configuration value: %s", err)
		}
		fmt.Printf("max-iterations: %d\n", i)

		return nil
	}
	onTerminate := func(s os.Signal) error {
		return nil
	}
	injectValidService()

	roverlib.Run(main, onTerminate)
}

// Test and access all service values
func TestValidProgramWithServiceAccess(t *testing.T) {
	main := func(s roverlib.Service, config *roverlib.ServiceConfiguration) error {
		fmt.Printf("Stated as service %s version %s\n", *s.Name, *s.Version)
		for _, input := range s.Inputs {
			for _, stream := range input.Streams {
				fmt.Printf("Input stream %s from %s\n", *stream.Name, *input.Service)
			}
		}
		for _, output := range s.Outputs {
			fmt.Printf("Output stream %s\n", *output.Name)
		}

		return nil
	}
	onTerminate := func(s os.Signal) error {
		fmt.Printf("Terminated with signal: %s\n", s)
		return nil
	}
	injectValidService()

	roverlib.Run(main, onTerminate)
}
