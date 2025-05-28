//
// Functionality and methods for working with streams to read and write data to/from other services
// all methods are meant to be used in the user program and are implemented on the "Service" struct that is passed to the user program main function
//

package roverlib

import (
	"fmt"
	"strings"

	rovercom "github.com/VU-ASE/rovercom/packages/go/outputs"
	"github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// Map of all already handed out streams to the user program (to preserve singletons)
var writeStreams = make(map[string]*WriteStream)
var readStreams = make(map[string]*ReadStream)

type serviceStream struct {
	// The socket that this stream is connected to
	address  string       // zmq address
	socket   *zmq4.Socket // can be nil, when lazy loading
	sockType zmq4.Type
	// Amount of bytes read/written so far
	bytes int
}

type WriteStream struct {
	stream serviceStream
}

type ReadStream struct {
	stream serviceStream
}

// Get a stream that you can write to (i.e. an output stream).
// This function panics if the stream does not exist, because fetching a non-existent stream should always terminate to avoid undefined behavior.
func (s *Service) GetWriteStream(name string) *WriteStream {
	// Is this stream already handed out?
	if stream, ok := writeStreams[name]; ok {
		return stream
	}

	// Does this stream exist?
	for _, output := range s.Outputs {
		if *output.Name == name {
			// ZMQ wants to bind write streams to tcp://*:port addresses, so if roverd gave us a localhost, we need to change it to *
			address := strings.Replace(*output.Address, "localhost", "*", 1)

			// Create a new stream
			stream := &serviceStream{
				address:  address,
				sockType: zmq4.PUB,
			}
			res := &WriteStream{stream: *stream}
			writeStreams[name] = res
			return res
		}
	}

	log.Error().Msgf("Output stream %s does not exist. Update your program code or service.yaml", name)
	return nil
}

// Get a stream that you can read from (i.e. an input stream).
// This function panics if the stream does not exist, because fetching a non-existent stream should always terminate to avoid undefined behavior.
func (s *Service) GetReadStream(service string, name string) *ReadStream {
	streamName := fmt.Sprintf("%s-%s", service, name)
	// Is this stream already handed out?
	if stream, ok := readStreams[streamName]; ok {
		return stream
	}

	// Does this stream exist?
	for _, input := range s.Inputs {
		if *input.Service == service {
			for _, stream := range input.Streams {
				if *stream.Name == name {
					// Create a new stream
					stream := &serviceStream{
						address:  *stream.Address,
						sockType: zmq4.SUB,
					}
					res := &ReadStream{stream: *stream}
					readStreams[streamName] = res
					return res
				}
			}
		}
	}

	log.Error().Msgf("Input stream %s does not exist. Update your program code or service.yaml", streamName)
	return nil
}

// Initial setup of the stream (done lazily, on the first read)
func (s *ReadStream) init() error {
	// Already initialized
	if s.stream.socket != nil {
		return nil
	}

	// Create a new socket
	socket, err := zmq4.NewSocket(s.stream.sockType)
	if err != nil {
		return fmt.Errorf("Failed to create read socket at %s: %w", s.stream.address, err)
	}
	err = socket.Connect(s.stream.address)
	if err != nil {
		return fmt.Errorf("Failed to connect read socket to %s: %w", s.stream.address, err)
	}
	err = socket.SetSubscribe("")
	if err != nil {
		return fmt.Errorf("Failed to set subscription on read socket: %w", err)
	}
	s.stream.socket = socket
	s.stream.bytes = 0
	return nil
}

// Initial setup of the stream (done lazily, on the first read)
func (s *WriteStream) init() error {
	// Already initialized
	if s.stream.socket != nil {
		return nil
	}

	// Create a new socket
	socket, err := zmq4.NewSocket(s.stream.sockType)
	if err != nil {
		return fmt.Errorf("Failed to create write socket at %s: %w", s.stream.address, err)
	}
	err = socket.Bind(s.stream.address)
	if err != nil {
		return fmt.Errorf("Failed to bind write socket to %s: %w", s.stream.address, err)
	}
	s.stream.socket = socket
	s.stream.bytes = 0
	return nil
}

// Write byte data to the stream
func (s *WriteStream) WriteBytes(data []byte) error {
	if s.stream.socket == nil {
		err := s.init()
		if err != nil {
			return err
		}
	}

	// Check if the socket is writable
	if s.stream.sockType != zmq4.PUB {
		return fmt.Errorf("Cannot write to a read-only stream")
	}

	// Write the data
	_, err := s.stream.socket.SendBytes(data, 0)
	if err != nil {
		return fmt.Errorf("Failed to write to stream: %w", err)
	}
	s.stream.bytes += len(data)
	return nil
}

// Read byte data from the stream
func (s *ReadStream) ReadBytes() ([]byte, error) {
	if s.stream.socket == nil {
		err := s.init()
		if err != nil {
			return nil, err
		}
	}

	// Check if the socket is readable
	if s.stream.sockType != zmq4.SUB {
		return nil, fmt.Errorf("Cannot read from a write-only stream")
	}

	// Read the data
	data, err := s.stream.socket.RecvBytes(0)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from stream: %w", err)
	}
	s.stream.bytes += len(data)
	return data, nil
}

// Write a rovercom sensor output message to the stream
func (s *WriteStream) Write(output *rovercom.SensorOutput) error {
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
func (s *ReadStream) Read() (*rovercom.SensorOutput, error) {
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
