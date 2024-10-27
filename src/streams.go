//
// Functionality and methods for working with streams to read and write data to/from other services
// all methods are meant to be used in the user program and are implemented on the "Service" struct that is passed to the user program main function
//

package roverlib

import (
	"fmt"
	"log"

	rovercom "github.com/VU-ASE/rovercom/packages/go/outputs"
	"github.com/pebbe/zmq4"
	"google.golang.org/protobuf/proto"
)

// Map of all already handed out streams to the user program (to preserve singletons)
var streams = make(map[string]*ServiceStream)

type ServiceStream struct {
	// The socket that this stream is connected to
	address  string       // zmq address
	socket   *zmq4.Socket // can be nil, when lazy loading
	sockType zmq4.Type
	// Amount of bytes read/written so far
	bytes int
}

// Get a stream that you can write to (i.e. an output stream).
// This function panics if the stream does not exist, because fetching a non-existent stream should always terminate to avoid undefined behavior.
func (s *Service) GetWriteStream(name string) *ServiceStream {
	// Is this stream already handed out?
	if stream, ok := streams[name]; ok {
		return stream
	}

	// Does this stream exist?
	for _, output := range s.Outputs {
		if *output.Name == name {
			// Create a new stream
			stream := &ServiceStream{
				address:  *output.Address,
				sockType: zmq4.PUB,
			}
			streams[name] = stream
			return stream
		}
	}

	log.Fatalf("Output stream %s does not exist. Update your program code or service.yaml", name)
	return nil
}

// Get a stream that you can read from (i.e. an input stream).
// This function panics if the stream does not exist, because fetching a non-existent stream should always terminate to avoid undefined behavior.
func (s *Service) GetReadStream(service string, name string) *ServiceStream {
	streamName := fmt.Sprintf("%s-%s", service, name)
	// Is this stream already handed out?
	if stream, ok := streams[streamName]; ok {
		return stream
	}

	// Does this stream exist?
	for _, input := range s.Inputs {
		if *input.Service == service {
			for _, stream := range input.Streams {
				if *stream.Name == name {
					// Create a new stream
					stream := &ServiceStream{
						address:  *stream.Address,
						sockType: zmq4.SUB,
					}
					streams[streamName] = stream
					return stream
				}
			}
		}
	}

	log.Fatalf("Input stream %s does not exist. Update your program code or service.yaml", streamName)
	return nil
}

// Initial setup of the stream (done lazily, on the first read or write)
func (s *ServiceStream) init() error {
	// Already initialized
	if s.socket != nil {
		return nil
	}

	// Create a new socket
	socket, err := zmq4.NewSocket(s.sockType)
	if err != nil {
		return err
	}
	err = socket.Connect(s.address)
	if err != nil {
		return err
	}
	s.socket = socket
	s.bytes = 0
	return nil
}

// Write byte data to the stream
func (s *ServiceStream) WriteBytes(data []byte) error {
	if s.socket == nil {
		err := s.init()
		if err != nil {
			return err
		}
	}

	// Check if the socket is writable
	if s.sockType != zmq4.PUB {
		return fmt.Errorf("Cannot write to a read-only stream")
	}

	// Write the data
	_, err := s.socket.SendBytes(data, 0)
	if err != nil {
		return err
	}
	s.bytes += len(data)
	return nil
}

// Read byte data from the stream
func (s *ServiceStream) ReadBytes() ([]byte, error) {
	if s.socket == nil {
		err := s.init()
		if err != nil {
			return nil, err
		}
	}

	// Check if the socket is readable
	if s.sockType != zmq4.SUB {
		return nil, fmt.Errorf("Cannot read from a write-only stream")
	}

	// Read the data
	data, err := s.socket.RecvBytes(0)
	if err != nil {
		return nil, err
	}
	s.bytes += len(data)
	return data, nil
}

// Write a rovercom sensor output message to the stream
func (s *ServiceStream) Write(output *rovercom.SensorOutput) error {
	if output == nil {
		return fmt.Errorf("Cannot write nil output")
	}

	// Marshal (convert to over-the-wire format)
	buf, err := proto.Marshal(output)
	if err != nil {
		return err
	}

	// Write the data
	return s.WriteBytes(buf)
}

// Read a rovercom sensor output message from the stream
// (you will need to switch on the returned message type to cast it to the correct type)
func (s *ServiceStream) Read() (*rovercom.SensorOutput, error) {
	// Read the data
	buf, err := s.ReadBytes()
	if err != nil {
		return nil, err
	}

	// Unmarshal (convert from over-the-wire format)
	output := &rovercom.SensorOutput{}
	err = proto.Unmarshal(buf, output)
	if err != nil {
		return nil, err
	}

	return output, nil
}
