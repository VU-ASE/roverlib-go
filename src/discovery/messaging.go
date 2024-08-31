package discovery

import (
	"fmt"
	"time"

	pb_core_messages "github.com/VU-ASE/rovercom/packages/go/core"
	"github.com/VU-ASE/roverlib/src/runner"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

//
// Exposed package functions to send and receive messages to and from the core
//

func SendRequestToCore(message *pb_core_messages.CoreMessage, core runner.CoreInfo) (*pb_core_messages.CoreMessage, error) {
	// Get the address to send to
	addr := core.RepReqAddress

	// create a zmq client socket to the core
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("could not open ZMQ connection to core: %s", err)
	}
	defer client.Close()
	err = client.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to core: %s", err)
	}

	// convert the message to bytes
	msgBytes, err := proto.Marshal(message)
	if err != nil {
		log.Err(err).Msg("Error marshalling protobuf message")
		return nil, err
	}

	// send registration to the core
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
				log.Warn().Msgf("Still waiting for response from core. Are you sure the core is running and available at '%s'?", addr)
			} else {
				log.Info().Msg("Waiting for response from core")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// wait for the response
	msg, err := client.RecvBytes(0)
	responseReceived = true
	if err != nil {
		return nil, err
	}

	// parse the response
	parsedMsg := pb_core_messages.CoreMessage{}
	err = proto.Unmarshal(msg, &parsedMsg)
	if err != nil {
		return nil, err
	}

	return &parsedMsg, nil
}
