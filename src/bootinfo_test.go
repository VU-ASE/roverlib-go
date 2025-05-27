package roverlib_test

import (
	"fmt"
	"os"
	"testing"
	"reflect"

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
				"type": "number",
				"tunable": true,
				"value": 100
			},
			{
				"name": "speed",
				"type": "number",
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

func injectInvalidService(variant string) {
	switch variant {
	case "no-fields":
		os.Setenv("ASE_SERVICE", `{}`)
	case "no-name":
		os.Setenv("ASE_SERVICE", `{"version": "1.0.1"}`)
	case "no-version":
		os.Setenv("ASE_SERVICE", `{"name": "controller"}`)
	case "no-inputs":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "1.0.1"}`)
	case "no-outputs":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "1.0.1", "inputs": []}`)
	case "no-configuration":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "1.0.1", "inputs": [], "outputs": []}`)
	case "no-tuning":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "1.0.1", "inputs": [], "outputs": [], "configuration": []}`)
	case "malformed-json":
		os.Setenv("ASE_SERVICE", `{"not": "correct json",}`)
	case "invalid-name":
		os.Setenv("ASE_SERVICE", `{"name": "Controller!", "version": "1.0.1", "inputs": [], "outputs": [], "configuration": [], "tuning": {}}`)
	case "invalid-version":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "v1.0", "inputs": [], "outputs": [], "configuration": [], "tuning": {}}`)
	case "invalid-alias":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "as": "Controller", "version": "1.0.1", "inputs": [], "outputs": [], "configuration": [], "tuning": {}}`)
	case "invalid-input-type":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "1.0.1", "inputs": "not-an-array", "outputs": [], "configuration": [], "tuning": {}}`)
	case "extra-field":
		os.Setenv("ASE_SERVICE", `{"name": "controller", "version": "1.0.1", "inputs": [], "outputs": [], "configuration": [], "tuning": {}, "extra": true}`)
	case "missing-input-stream":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1",
			"inputs": [{"service": "imaging", "streams": [{}]}],
			"outputs": [], "configuration": [], "tuning": {}
		  }`)
	case "missing-input-service":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1",
			"inputs": [{"service": "", "streams": [{"name": "track_data", "address": "tcp://unix:7890"}]}],
			"outputs": [], "configuration": [], "tuning": {}
		  }`)
	case "missing-output-name":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1",
			"inputs": [],
			"outputs": [{"name": "", "address": "tcp://unix:7882"}],
			"configuration": [], "tuning": {}
		  }`)
	case "missing-output-address":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1",
			"inputs": [],
			"outputs": [{"name": "motor_movement", "address": ""}],
			"configuration": [], "tuning": {}
		  }`)
	case "non-boolean":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1",
			"inputs": [], "outputs": [], "configuration": [],
			"tuning": {"enabled": "yes", "address": "tcp://unix:8829"}
		  }`)
	case "invalid-config-number":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1", "inputs": [], "outputs": [],
			"configuration": [{"name": "option", "type": "integer", "tunable": true, "value": 5}],
			"tuning": {"enabled": true, "address": "tcp://unix:8829"}
		  }`)
	case "type-mismatch":
		os.Setenv("ASE_SERVICE", `{
			"name": "controller", "version": "1.0.1", "inputs": [], "outputs": [],
			"configuration": [{"name": "option", "type": "string", "tunable": true, "value": 123}],
			"tuning": {"enabled": true, "address": "tcp://unix:8829"}
		  }`)
	}
}

// TESTING VALID BOOTSPEC

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
	var gotMaxIterations float64
	var gotSpeed float64
	var gotLogLevel string

	main := func(s roverlib.Service, config *roverlib.ServiceConfiguration) error {
		// Test float access
		i, err := config.GetFloat("max-iterations")
		if err != nil {
			t.Errorf("Failed to get number configuration value: %s", err)
		}
		
		fmt.Printf("max-iterations: %f\n", i)
		// Test SAFE float access and save the values to be asserted later
		i, err = config.GetFloatSafe("max-iterations")
		if err != nil {
			t.Errorf("Failed to get number configuration value: %s", err)
		}
		gotMaxIterations = i

		i, err = config.GetFloat("speed")
		if err != nil {
			t.Errorf("Failed to get number configuration value: %s", err)
		}
		gotSpeed = i

		j, err := config.GetString("log-level")
		if err != nil {
			t.Errorf("Failed to get string configuration value: %s", err)
		}
		gotLogLevel = j
		return nil
	}
	onTerminate := func(s os.Signal) error {
		return nil
	}

	injectValidService()
	roverlib.Run(main, onTerminate)

	// check if the values are matching with the expected ones
	if gotMaxIterations != 100 {
		t.Errorf("Expected max-iterations to be 100, got %f\n", gotMaxIterations)
	} else {
		t.Log("CORRECT max-iterations\n")
	}
	if gotSpeed != 1.5 {
		t.Errorf("Expected speed to be 1.5, got %f\n", gotSpeed)
	} else {
		fmt.Print("CORRECT speed\n")
	}
	if gotLogLevel != "debug" {
		t.Errorf("Expected log-level to be debug, got %s\n", gotLogLevel)
	} else {
		fmt.Print("CORRECT log-level\n")
	}
}

// Test and access all service values
func TestValidProgramWithServiceAccess(t *testing.T) {
	// Define the expected input and output structures
	type inputDefinition struct { sercie, name, address string }
	type outputDefinition struct { name, address string }
	// Define arrays for the inputs and outputs we will get from the service
	var gotInputs []inputDefinition
	var gotOutputs []outputDefinition

	main := func(s roverlib.Service, config *roverlib.ServiceConfiguration) error {
		// Store the inputs and outputs in the arrays
		fmt.Printf("Stated as service %s version %s\n", *s.Name, *s.Version)
		for _, input := range s.Inputs {
			for _, stream := range input.Streams {
				gotInputs = append(gotInputs, inputDefinition{
					sercie: *input.Service,
					name:   *stream.Name,
					address: *stream.Address,
				})
			}
		}
		for _, output := range s.Outputs {
			gotOutputs = append(gotOutputs, outputDefinition{
				name:    *output.Name,
				address: *output.Address,
			})
		}

		return nil
	}
	onTerminate := func(s os.Signal) error {
		fmt.Printf("Terminated with signal: %s\n", s)
		return nil
	}

	// inject valid service and run the program
	injectValidService()
	roverlib.Run(main, onTerminate)

	// Create a structure for the expected inputs and outputs
	// and compare them with the ones we got from the service
	wantInputs := []inputDefinition{
		{"imaging", "track_data", "tcp://unix:7890"},
		{"imaging", "debug_info", "tcp://unix:7891"},
		{"navigation", "location_data", "tcp://unix:7892"},
	}

	wantOutputs := []outputDefinition{
		{"motor_movement", "tcp://unix:7882"},
		{"sensor_data", "tcp://unix:7883"},
	}

	if !reflect.DeepEqual(gotInputs, wantInputs) {
		t.Errorf("Expected inputs to be %+v, got %+v\n", wantInputs, gotInputs)
	} else {
		t.Log("CORRECT inputs\n")
	}
	if !reflect.DeepEqual(gotOutputs, wantOutputs) {
		t.Errorf("Expected outputs to be %+v, got %+v\n", wantOutputs, gotOutputs)
	} else {
		t.Log("CORRECT outputs\n")
	}
}



// TESTING INVALID BOOTSPEC

// Test all but 1 of the invalid bootspecs, If Run panics then the test passes
// If Run does not panic then the test fails
func TestInvalidProgram(t *testing.T) {
	bootspecs := []string{
		"no-name",
		"no-version",
		"no-inputs",
		"no-outputs",
		"no-configuration",
		"no-tuning",
		"malformed-json",
		"invalid-name",
		"invalid-version",
		"invalid-alias",
		"invalid-input-type",
		"extra-field",
		"missing-input-stream",
		"missing-input-service",
		"missing-output-name",
		"missing-output-address",
		"non-boolean",
		// "invalid-config-number",		Run never checks the type of the config so this test doesn't invoke a panic
		"type-mismatch",
	}

	main := func(s roverlib.Service, config *roverlib.ServiceConfiguration) error {
		fmt.Printf("Stated as service %s version %s\n", *s.Name, *s.Version)
		
		return nil
	}
	onTerminate := func(s os.Signal) error {
		fmt.Printf("Terminated with signal: %s\n", s)
		return nil
	}

	for _, bootspec := range bootspecs {
		bootspec := bootspec
		t.Run(bootspec, func(t *testing.T) {
			t.Logf("Testing invalid bootspec: %q\n", bootspec)

			defer func() {
				if r:= recover(); r == nil {
					t.Fatalf("expected Run to panic, but it did not: %v", r)
				}	
			}()
			injectInvalidService(bootspec)
			roverlib.Run(main, onTerminate)
		
		})
	}

}

func TestInvalidConfigType(t *testing.T) {
	injectInvalidService("invalid-config-number")
	
	// testing with Unmarshal and GetFloat whether they will fail for an invalid type in the configuration
	service, err := roverlib.UnmarshalService([]byte(os.Getenv("ASE_SERVICE")))
	if err != nil {
		t.Fatalf("Failed to unmarshal service definition in ASE_SERVICE: %s", err)
	}
	config := roverlib.NewServiceConfiguration(service)
	_, err = config.GetFloat("option")
	if err == nil {
		t.Fatalf("Expected error when getting invalid config type, but got none")
	}
}
