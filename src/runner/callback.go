package runner

import (
	"fmt"
	"strings"

	pb_output_messages "github.com/VU-ASE/rovercom/packages/go/outputs"
	"github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Stream interface {
	Recv() (*pb_output_messages.SensorOutput, error)
	Send(msg *pb_output_messages.SensorOutput) error
	Close() error
}

//
// This file contains the information that is passed to the service that is run. The service can then use this information to connect to the input streams or send info to output streams.
//

type Service struct {
	Name    string          // the name of the service
	Pid     int             // the pid of the service
	Inputs  []ServiceInput  // the dependencies of the service, already resolved
	Outputs []ServiceOutput // the outputs of this service
}

type ServiceInput struct {
	Service string // the name of the service that this service depends on
	Stream  string // the name of the stream that the running service needs from the dependency service
	Address string // the address of the output that this service needs from the dependency
}

type ServiceOutput struct {
	Name    string // the name of the output stream
	Address string // the address of the output stream that can be written to
}

// ZMQ stream that can be used to read data
type InputStream struct {
	socket *zmq4.Socket
}

// Close the ZMQ stream
func (stream *InputStream) Close() error {
	if stream.socket != nil {
		return stream.socket.Close()
	}
	return nil
}

// Receive raw bytes from a ZMQ stream
func (stream *InputStream) RecvBytes() (string, error) {
	return stream.socket.Recv(0)
}

// Receive a rovercom output message from a ZMQ stream
func (stream *InputStream) Recv() (*pb_output_messages.SensorOutput, error) {
	data, err := stream.RecvBytes()
	if err != nil {
		return nil, err
	}

	msg := &pb_output_messages.SensorOutput{}
	err = proto.Unmarshal([]byte(data), msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// ZMQ stream that can be used to write data
type OutputStream struct {
	socket *zmq4.Socket
}

// Close the ZMQ stream
func (stream *OutputStream) Close() error {
	if stream.socket != nil {
		return stream.socket.Close()
	}
	return nil
}

// Send raw bytes to a ZMQ stream
func (stream *OutputStream) SendBytes(data string) error {
	_, err := stream.socket.Send(data, 0)
	return err
}

// Send a rovercom output message to a ZMQ stream
func (stream *OutputStream) Send(msg *pb_output_messages.SensorOutput) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return stream.SendBytes(string(data))
}

// For now, we only replace * with localhost for zmq, but more modifications can be added later
func rewriteInputAddress(addr string) string {
	depAddr := strings.ReplaceAll(addr, "*", "localhost")
	if depAddr != addr {
		log.Debug().Str("old", addr).Str("new", depAddr).Msg("Rewrote dependency address for own consumption")
	}
	return depAddr
}

// Utiliy function to get the address of a dependency
func (service Service) GetInputAddress(serviceName string, streamName string) (string, error) {
	for _, input := range service.Inputs {
		if strings.EqualFold(serviceName, input.Service) && strings.EqualFold(input.Stream, streamName) {
			return rewriteInputAddress(input.Address), nil
		}
	}
	return "", fmt.Errorf("strean '%s.%s' not found. Are you sure it is exposed by %s?", serviceName, streamName, serviceName)
}

// For now, we only replace localhost with * for zmq, but more modifications can be added later
func rewriteOutputAddress(addr string) string {
	repAddr := strings.ReplaceAll(addr, "localhost", "*")
	if repAddr != addr {
		log.Debug().Str("old", addr).Str("new", repAddr).Msg("Rewrote output address for own consumption")
	}
	return repAddr
}

// Utility function to get the address of your own output
func (service Service) GetOutputAddress(streamName string) (string, error) {
	for _, output := range service.Outputs {
		if strings.EqualFold(streamName, output.Name) {
			return rewriteOutputAddress(output.Address), nil
		}
	}
	return "", fmt.Errorf("output '%s' not found. Was it defined in service.yaml?", streamName)
}

// Information about the core. This struct has useful methods implemented to repeat the same operations on the dependencies.
type CoreInfo struct {
	RepReqAddress    string // the req/rep address of the core
	BroadcastAddress string // the public broadcast address of the core
}

// Utility function to get a list of all services running on the core
// func (core CoreInfo) GetAllServices() (*pb_core_messages.ServiceList, error) {
// 	// return getServiceList(core.RepReqAddress)
// 	return nil, nil // todo
// }

// // Utility function to get the latest tuning state from the core
// func (core CoreInfo) GetTuningState() (*pb_core_messages.TuningState, error) {
// 	// return getTuningState(core.RepReqAddress)
// 	return nil, nil // todo
// }

func (input ServiceInput) ToStream() (*InputStream, error) {
	// Initialize the socket
	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}

	// Connect to the address
	err = socket.Connect(rewriteInputAddress(input.Address))
	if err != nil {
		return nil, err
	}

	// Subscribe to all messages
	err = socket.SetSubscribe("")
	if err != nil {
		return nil, err
	}

	return &InputStream{socket: socket}, nil
}

func (output ServiceOutput) ToStream() (*OutputStream, error) {
	// Initialize the socket
	socket, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		return nil, err
	}

	// Bind to the address
	err = socket.Bind(rewriteOutputAddress(output.Address))
	if err != nil {
		return nil, err
	}

	return &OutputStream{socket: socket}, nil
}

// Get a stream to read from, do not forget to close it
func (service Service) GetInputStream(serviceName string, streamName string) (*InputStream, error) {
	for _, input := range service.Inputs {
		if strings.EqualFold(serviceName, input.Service) && strings.EqualFold(input.Stream, streamName) {
			return input.ToStream()
		}
	}
	return nil, fmt.Errorf("stream '%s.%s' not found. Are you sure it is exposed by %s?", serviceName, streamName, serviceName)
}

// Get a stream to write to, do not forget to close it
func (service Service) GetOutputStream(streamName string) (*OutputStream, error) {
	for _, output := range service.Outputs {
		if strings.EqualFold(streamName, output.Name) {
			return output.ToStream()
		}
	}
	return nil, fmt.Errorf("output '%s' not found. Was it defined in service.yaml?", streamName)
}

// // Utility function to easily read values from the tuning state
// func GetTuningInt(key string, tuningState *pb_core_messages.TuningState) (int, error) {
// 	if tuningState == nil {
// 		return 0, fmt.Errorf("tuning state is nil")
// 	}

// 	// Iterate over all the tuning state values
// 	for _, tuningValue := range tuningState.DynamicParameters {
// 		val := tuningValue.GetInt()
// 		if val != nil && val.Key == key {
// 			return int(val.Value), nil
// 		}
// 	}

// 	return 0, fmt.Errorf("key '%s' not found in tuning state", key)
// }

// // Utility function to easily read values from the tuning state
// func GetTuningString(key string, tuningState *pb_core_messages.TuningState) (string, error) {
// 	if tuningState == nil {
// 		return "", fmt.Errorf("tuning state is nil")
// 	}

// 	// Iterate over all the tuning state values
// 	for _, tuningValue := range tuningState.DynamicParameters {
// 		val := tuningValue.GetString_()
// 		if val != nil && val.Key == key {
// 			return val.Value, nil
// 		}
// 	}

// 	return "", fmt.Errorf("key '%s' not found in tuning state", key)
// }

// // Utility function to easily read values from the tuning state
// func GetTuningFloat(key string, tuningState *pb_core_messages.TuningState) (float32, error) {
// 	if tuningState == nil {
// 		return 0, fmt.Errorf("tuning state is nil")
// 	}

// 	// Iterate over all the tuning state values
// 	for _, tuningValue := range tuningState.DynamicParameters {
// 		val := tuningValue.GetFloat()
// 		if val != nil && val.Key == key {
// 			return val.Value, nil
// 		}
// 	}

// 	return 0, fmt.Errorf("key '%s' not found in tuning state", key)
// }
