package servicerunner

import (
	"fmt"

	pb_systemmanager_messages "github.com/VU-ASE/pkg-CommunicationDefinitions/v2/packages/go/systemmanager"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

//
// Exposed package functions to send and receive messages to and from the system manager
//

func SendRequestToSystemManager(message *pb_systemmanager_messages.SystemManagerMessage) (*pb_systemmanager_messages.SystemManagerMessage, error) {
	// Get the address to send to
	addr, err := getSystemManagerRepReqAddress()
	if err != nil {
		return nil, err
	}

	// create a zmq client socket to the system manager
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("Could not open ZMQ connection to system manager: %s", err)
	}
	defer client.Close()
	err = client.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to system manager: %s", err)
	}

	// convert the message to bytes
	msgBytes, err := proto.Marshal(message)
	if err != nil {
		log.Err(err).Msg("Error marshalling protobuf message")
		return nil, err
	}

	// send registration to the system manager
	_, err = client.SendBytes(msgBytes, 0)
	if err != nil {
		return nil, err
	}

	// wait for the response
	msg, err := client.RecvBytes(0)
	if err != nil {
		return nil, err
	}

	// parse the response
	parsedMsg := pb_systemmanager_messages.SystemManagerMessage{}
	err = proto.Unmarshal(msg, &parsedMsg)
	if err != nil {
		return nil, err
	}

	return &parsedMsg, nil
}
