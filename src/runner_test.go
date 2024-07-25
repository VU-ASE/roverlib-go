package servicerunner

import (
	"testing"

	pb_systemmanager_messages "github.com/VU-ASE/pkg-CommunicationDefinitions/v2/packages/go/systemmanager"
	"github.com/rs/zerolog/log"
)

// Show logs
func TestLogging(t *testing.T) {
	setupLogging(false, "", serviceDefinition{
		Name: "MyTestService",
	})
	log.Info().Msg("This is a test log")
}

func TestBasicTuningUpdate(t *testing.T) {
	serviceDef := serviceDefinition{
		Name: "TuningStateService",
		Options: []option{
			{
				Name:         "param1",
				DefaultValue: "4",
				Mutable:      true,
				Type:         "int",
			},
			{
				Name:         "param2",
				DefaultValue: "teststr",
				Mutable:      false,
				Type:         "string",
			},
		},
	}

	// Create initial tuning state
	initialTuning, err := convertOptionsToTuningState(serviceDef.Options)
	if err != nil {
		t.Errorf("Error converting options to tuning state: %v", err)
	}

	// Create an updated tuning state (as would be received over the air)
	receivedTuning := &pb_systemmanager_messages.TuningState{
		DynamicParameters: []*pb_systemmanager_messages.TuningState_Parameter{
			{
				Parameter: &pb_systemmanager_messages.TuningState_Parameter_Int{
					Int: &pb_systemmanager_messages.TuningState_Parameter_IntParameter{
						Key:   "param1",
						Value: 10,
					},
				},
			},
			{
				Parameter: &pb_systemmanager_messages.TuningState_Parameter_String_{
					String_: &pb_systemmanager_messages.TuningState_Parameter_StringParameter{
						Key:   "param2",
						Value: "abc",
					},
				},
			},
			{
				Parameter: &pb_systemmanager_messages.TuningState_Parameter_String_{
					String_: &pb_systemmanager_messages.TuningState_Parameter_StringParameter{
						Key:   "param3",
						Value: "you should not see this",
					},
				},
			},
		},
	}

	// Merge the states, create a new state
	newTuning := createUpdatedTuningState(initialTuning, receivedTuning, serviceDef.Options)

	// Expect that param1 is updated, param2 is not updated, and param3 is not present
	if keyExists("param3", "string", newTuning) {
		t.Errorf("Unexpected key 'param3' in new tuning state")
	}
	if !keyExists("param1", "int", newTuning) {
		t.Errorf("Expected key 'param1' in new tuning state")
	}
	if !keyExists("param2", "string", newTuning) {
		t.Errorf("Expected key 'param2' in new tuning state")
	}
	if len(newTuning.DynamicParameters) != 2 {
		t.Errorf("Expected 2 dynamic parameters, got %d", len(newTuning.DynamicParameters))
	}

	// Expect that param1 is updated to 10, and param2 is updated to "abc"
	for _, param := range newTuning.DynamicParameters {
		switch param.Parameter.(type) {
		case *pb_systemmanager_messages.TuningState_Parameter_Int:
			if param.GetInt().Value != 10 {
				t.Errorf("Expected param1 to be 10, got %d", param.GetInt().Value)
			}
		case *pb_systemmanager_messages.TuningState_Parameter_String_:
			if param.GetString_().Value != "teststr" {
				t.Errorf("Expected param2 to be 'teststr', got %s", param.GetString_().Value)
			}
		}
	}
}
