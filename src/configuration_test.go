package roverlib

import (
	//"fmt"
	"sync"
	"testing"
	"reflect"

	//roverlib "github.com/VU-ASE/roverlib-go/src"
)

// helper: make a tiny Service with 1 float (tunable), 1 string entry (not tunable)
func sampleService() Service {
	num := Number
	str := String
	tTrue, tFalse := true, false
	name1, name2 := "config1", "config2"
	valConfig1 := 3.14
	valConfig2 := "auto"
	return Service{
		Configuration: []Configuration{
			{Name: &name1, Type: &num,   Tunable: &tTrue,  Value: &Value{Double: &valConfig1}},
			{Name: &name2, Type: &str,   Tunable: &tFalse, Value: &Value{String: &valConfig2}},
		},
	}
}

// Checks that NewServiceConfiguration correctly builds the maps for float, string, and tunable options
func TestNewServiceConfigurationBuildsMaps(t *testing.T) {
	cfg := NewServiceConfiguration(sampleService())
	wantFloats  := map[string]float64{"config1": 3.14}
	wantStrings := map[string]string{"config2": "auto"}
	wantTunable := map[string]bool{"config1": true, "config2": false}

	if !reflect.DeepEqual(cfg.floatOptions, wantFloats) {
		t.Fatalf("floatOptions = %#v, want %#v", cfg.floatOptions, wantFloats)
	}
	if !reflect.DeepEqual(cfg.stringOptions, wantStrings) {
		t.Fatalf("stringOptions = %#v, want %#v", cfg.stringOptions, wantStrings)
	}
	if !reflect.DeepEqual(cfg.tunable, wantTunable) {
		t.Fatalf("tunable = %#v, want %#v", cfg.tunable, wantTunable)
	}
}

// Tests the getters for float and string values, both happy and sad paths
// the first and last if statements are happy paths, the middle one is a sad path
func TestGettersHappyAndSad(t *testing.T) {
	cfg := NewServiceConfiguration(sampleService())

	if v, _ := cfg.GetFloat("config1"); v != 3.14 {
		t.Fatalf("GetFloat(config1) = %v, want 3.14", v)
	}
	if _, err := cfg.GetFloat("config2"); err == nil {
		t.Fatalf("expected error for GetFloat on string key")
	}
	if s, _ := cfg.GetString("config2"); s != "auto" {
		t.Fatalf("GetString(config2) = %q, want \"auto\"", s)
	}
}

// Tests the safe getters for float, it spawns 20 gorutines that will simultaneously read
// the same key using the GetFloatSafe method. The tests pass if no error is returned
func TestGetFloatSafeConcurrent(t *testing.T) {
	cfg := NewServiceConfiguration(sampleService())
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cfg.GetFloatSafe("config1"); err != nil {
				t.Errorf("concurrent GetFloatSafe returned %v", err)
			}
		}()
	}
	wg.Wait()
}

// Same as above, but for the string configuration option
func TestGetStringSafeConcurrent(t *testing.T) {
	cfg := NewServiceConfiguration(sampleService())
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := cfg.GetStringSafe("config2"); err != nil {
				t.Errorf("concurrent GetStringSafe returned %v", err)
			}
		}()
	}
	wg.Wait()
}

// Tests setting a float configuration on both the tunable and non-tunable options
func TestSetFloatTunable(t *testing.T) {
	cfg := NewServiceConfiguration(sampleService())
	cfg.setFloat("config1", 6.28)    // tunable → should update
	if v, _ := cfg.GetFloat("config1"); v != 6.28 {
		t.Fatalf("pi not updated: got %v", v)
	}
	cfg.setFloat("config2", 1.23)  // not tunable → ignore
	if _, err := cfg.GetFloat("config2"); err == nil {
		t.Fatalf("string option should not be in float map")
	}
}

// Tests getters for a missing key in the configuration
func TestGettersMissingKey(t *testing.T) {
	cfg := NewServiceConfiguration(sampleService())

	if _, err := cfg.GetFloat("missing"); err == nil {
		t.Fatalf("expected error for missing float key")
	}
	if _, err := cfg.GetString("missing"); err == nil {
		t.Fatalf("expected error for missing string key")
	}
}
