package roverlib

import (
	"os"

	pb_core "github.com/VU-ASE/rovercom/packages/go/core"
	"github.com/VU-ASE/roverlib/src/runner"
)

// The function that is called when a new tuning state is recevied
type OnTuningStateCallback func(ts *pb_core.TuningState)

// The main function to run
type MainCallback func(s runner.Service, c runner.CoreInfo, ts *pb_core.TuningState) error

// The function to call when the service is terminated or interrupted
type TerminationCallback func(s os.Signal)
