package servicerunner

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	pb_core_messages "github.com/VU-ASE/rovercom/packages/go/core"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// Listens on the broadcast channel for broadcasts from the core
func listenForTuningBroadcasts(onTuningState TuningStateCallbackFunction, sysmanBroadcastAddr string) error {
	// Subscribe to the core
	broadcast, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		return err
	}
	defer broadcast.Close()
	err = broadcast.Connect(sysmanBroadcastAddr)
	if err != nil {
		return err
	}
	err = broadcast.SetSubscribe("")
	if err != nil {
		return err
	}

	// main receiver loop
	for {
		// Receive the tuning state
		msg, err := broadcast.RecvBytes(0)
		if err != nil {
			return err
		}

		// NOTE: this is a best effor, we discard the message (and any errors) if we cannot parse it
		// Parse the tuning state
		parsedMsg := pb_core_messages.CoreMessage{}
		err = proto.Unmarshal(msg, &parsedMsg)
		if err != nil {
			continue
		}

		newTuningState := parsedMsg.GetTuningState()
		if newTuningState == nil {
			log.Warn().Msg("Received empty tuning state broadcast from core")
			continue
		}

		// Send the tuning state to the callback function
		log.Debug().Msgf("Received tuning state broadcast from core: %s", newTuningState.String())
		onTuningState(newTuningState)
	}
}

// This will combine the current state with the new state, and return the combined state
// note: it will edit the new state in placepb_core_messages
func createUpdatedTuningState(currentTuning *pb_core_messages.TuningState, receivedTuning *pb_core_messages.TuningState, serviceOptions []option) *pb_core_messages.TuningState {
	if currentTuning == nil || receivedTuning == nil {
		return nil
	}

	// Strip the received tuning state of any parameters that are marked as "mutable: false" in the service.yaml file, not present in the service options, or have a different type than the one in the service options
	// This is done to prevent the core from overwriting parameters that are not meant to be changed during runtime
	newTuningStateParams := receivedTuning.GetDynamicParameters()
	if newTuningStateParams == nil {
		return nil
	}
	newTuningStateParams = slices.DeleteFunc(newTuningStateParams, func(tuningParam *pb_core_messages.TuningState_Parameter) bool {
		// Does a parameter with the same key and type exist in the service options?
		for _, opt := range serviceOptions {
			// Does this parameter exist in the service options?
			if tuningParameterMatchesOption(tuningParam, opt) {
				return !opt.Mutable // delete if not mutable
			}
		}

		return true // delete anyway, as it is not present in the service options
	})
	receivedTuning.DynamicParameters = newTuningStateParams
	merged := mergeTuningStates(currentTuning, receivedTuning)
	// Delete all parameters that are not present in the service options
	merged.DynamicParameters = slices.DeleteFunc(merged.DynamicParameters, func(tuningParam *pb_core_messages.TuningState_Parameter) bool {
		for _, opt := range serviceOptions {
			if tuningParameterMatchesOption(tuningParam, opt) {
				return false
			}
		}
		return true
	})

	return merged
}

// This utility function will check if the type and name of a tuning parameter matches a given option as defined in the service.yaml file
// This is used to filter out parameters that are not being used by the service during runtime
func tuningParameterMatchesOption(tuningParam *pb_core_messages.TuningState_Parameter, opt option) bool {
	if tuningParam == nil {
		return false
	}

	// Get the key and type of the tuning parameter
	key, keyType := getKeyAndType(tuningParam)
	return key == opt.Name && keyType == opt.Type
}

func requestTuningState(sysmanReqRepAddr string) (*pb_core_messages.TuningState, error) {
	// create a zmq client socket to the core
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("could not open ZMQ connection to core: %s", err)
	}
	defer client.Close()
	log.Debug().Str("address", sysmanReqRepAddr).Msg("Connecting to core to fetch tuning state")
	err = client.Connect(sysmanReqRepAddr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to core: %s", err)
	}

	// create a request message
	reqMsg := pb_core_messages.CoreMessage{
		Msg: &pb_core_messages.CoreMessage_TuningStateRequest{
			TuningStateRequest: &pb_core_messages.TuningStateRequest{},
		},
	}

	// convert the message to bytes
	msgBytes, err := proto.Marshal(&reqMsg)
	if err != nil {
		log.Err(err).Msg("Error marshalling protobuf message")
		return nil, err
	}

	// send request to the core
	log.Info().Msg("Requesting tuning state from core")
	_, err = client.SendBytes(msgBytes, 0)
	if err != nil {
		return nil, err
	}

	responseReceived := false
	go func() {
		count := 0
		for {
			// print a idle message every 5 seconds, until were done
			if responseReceived {
				return
			}
			if (count) > 5 {
				log.Warn().Msg("Still waiting for tuning state response from core. Are you sure the core is running ?")
			} else {
				log.Info().Msg("Waiting for tuning state response from core")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// wait for a response from the core
	resBytes, err := client.RecvBytes(0)
	responseReceived = true

	// the response must be of type TuningState (see include/tuningstate.proto)
	// if not, we discard it: registration not successful
	log.Info().Msg("Received tuning state response from core")
	if err != nil {
		return nil, err
	}
	response := pb_core_messages.CoreMessage{}
	err = proto.Unmarshal(resBytes, &response)
	if err != nil {
		log.Err(err).Msg("Error unmarshalling tuning state protobuf message")
		return nil, err
	}
	responseState := response.GetTuningState()
	if responseState == nil {
		return nil, fmt.Errorf("received empty response from core")
	}

	return responseState, nil
}

// Used to convert the options present in the service.yaml file to a tuning state
func convertOptionsToTuningState(options []option) (*pb_core_messages.TuningState, error) {
	tuningState := pb_core_messages.TuningState{
		Timestamp:         uint64(time.Now().Unix()),
		DynamicParameters: make([]*pb_core_messages.TuningState_Parameter, 0),
	}
	for _, opt := range options {
		if opt.DefaultValue != "" {
			newParam := pb_core_messages.TuningState_Parameter{}
			switch opt.Type {
			case "string":
				newParam.Parameter = &pb_core_messages.TuningState_Parameter_String_{
					String_: &pb_core_messages.TuningState_Parameter_StringParameter{
						Key:   opt.Name,
						Value: opt.DefaultValue,
					},
				}
			case "int":
				intval, err := strconv.Atoi(opt.DefaultValue)
				if err != nil {
					return nil, fmt.Errorf("error converting default value of option '%s' to int: %s", opt.Name, err)
				}
				newParam.Parameter = &pb_core_messages.TuningState_Parameter_Int{
					Int: &pb_core_messages.TuningState_Parameter_IntParameter{
						Key:   opt.Name,
						Value: int64(intval),
					},
				}
			case "float":
				floatval, err := strconv.ParseFloat(opt.DefaultValue, 32)
				if err != nil {
					return nil, fmt.Errorf("error converting default value of option '%s' to float: %s", opt.Name, err)
				}
				newParam.Parameter = &pb_core_messages.TuningState_Parameter_Float{
					Float: &pb_core_messages.TuningState_Parameter_FloatParameter{
						Key:   opt.Name,
						Value: float32(floatval),
					},
				}
			default:
				return nil, fmt.Errorf("unknown type '%s' for option '%s'", opt.Type, opt.Name)
			}

			if newParam.Parameter != nil {
				tuningState.DynamicParameters = append(tuningState.DynamicParameters, &newParam)
			}
		}
	}
	return &tuningState, nil
}

func getUnsetOptions(tuningState *pb_core_messages.TuningState, options []option) []option {
	unsetOptions := make([]option, 0)
	for _, opt := range options {
		var err error = nil
		switch opt.Type {
		case "string":
			_, err = GetTuningString(opt.Name, tuningState)
		case "int":
			_, err = GetTuningInt(opt.Name, tuningState)
		case "float":
			_, err = GetTuningFloat(opt.Name, tuningState)
		default:
			log.Err(err).Msg("Error getting option from tuning state")
			err = fmt.Errorf("unknown type '%s' for option '%s'", opt.Type, opt.Name)
		}
		if err != nil {
			unsetOptions = append(unsetOptions, opt)
		}
	}
	return unsetOptions
}

// Parses a parameter from the tuning state and returns the key
func getKeyAndType(param *pb_core_messages.TuningState_Parameter) (string, string) {
	if param.GetString_() != nil {
		return param.GetString_().Key, "string"
	} else if param.GetInt() != nil {
		return param.GetInt().Key, "int"
	} else if param.GetFloat() != nil {
		return param.GetFloat().Key, "float"
	}
	return "", ""
}

// Check if a key with specific type exists in the tuning state
func keyExists(key string, keyType string, tuningState *pb_core_messages.TuningState) bool {
	if tuningState == nil {
		return false
	}
	for _, tuningValue := range tuningState.DynamicParameters {
		if tuningValue == nil {
			continue
		}

		if tuningValue.GetString_() != nil && tuningValue.GetString_().Key == key && keyType == "string" {
			return true
		} else if tuningValue.GetInt() != nil && tuningValue.GetInt().Key == key && keyType == "int" {
			return true
		} else if tuningValue.GetFloat() != nil && tuningValue.GetFloat().Key == key && keyType == "float" {
			return true
		}
	}
	return false
}

func mergeTuningStates(old *pb_core_messages.TuningState, new *pb_core_messages.TuningState) *pb_core_messages.TuningState {
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}
	merged := pb_core_messages.TuningState{
		Timestamp:         new.Timestamp,
		DynamicParameters: make([]*pb_core_messages.TuningState_Parameter, 0),
	}
	// Tuning states can be partial. This will make sure that all the old parameters are added to the merged tuning state,
	// even if they are not present in the received new tuning state
	for _, param := range old.DynamicParameters {
		key, keyType := getKeyAndType(param)
		if !keyExists(key, keyType, new) {
			merged.DynamicParameters = append(merged.DynamicParameters, param)
		}
	}
	// Add all the new parameters, overwriting any old ones
	merged.DynamicParameters = append(merged.DynamicParameters, new.DynamicParameters...)
	return &merged
}
