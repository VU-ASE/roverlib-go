package roverlib

import (
	"os"
)

// The user main function to run
type MainCallback func(
	s Service, // Basic information about the service being run, so that you know who you are
	config *ServiceConfiguration, // The configuration options for this service (can be tuned ota)
) error

// The function to call when the service is terminated or interrupted
type TerminationCallback func(s os.Signal) error
