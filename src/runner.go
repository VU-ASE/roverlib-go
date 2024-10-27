package roverlib

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	build_debug "runtime/debug"

	"github.com/pebbe/zmq4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// The core exposes two endpoints: a pub/sub endpoint for broadcasting service registration and a req/rep endpoint for registering services and resolving dependencies
// this struct is used to store the addresses of these endpoints

// Configures log level and output
func setupLogging(debug bool, outputPath string, service InjectedService) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	// Set up custom caller prefix
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		path := strings.Split(file, "/")
		// only take the last three elements of the path
		filepath := strings.Join(path[len(path)-3:], "/")
		return fmt.Sprintf("[%s] %s:%d", *service.Name, filepath, line)
	}
	outputWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = log.Output(outputWriter).With().Caller().Logger()
	if outputPath != "" {
		file, err := os.OpenFile(
			outputPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			panic(err)
		}
		log.Logger = zerolog.New(file).With().Timestamp().Caller().Logger()
		fmt.Printf("Logging to file %s\n", outputPath)
	}

	// Set log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logs enabled")
	}

	log.Info().Msg("Logger was set up")

	if debug {
		log.Debug().Msg("Listing dependencies of this binary: ")
		buildInfo, ok := build_debug.ReadBuildInfo()
		if !ok {
			log.Warn().Msg("Failed to read build info")
		} else {
			for _, dep := range buildInfo.Deps {
				s := fmt.Sprintf("  dep: %s@%s", dep.Path, dep.Version)
				log.Debug().Msg(s)
			}
		}
	}
}

// Start the program (main) and handle termination
func Run(main MainCallback, onTerminate TerminationCallback) {
	// Parse args
	debug := flag.Bool("debug", false, "show all logs (including debug)")
	output := flag.String("output", "", "path of the output file to log to")
	flag.Parse()

	// Catch sigterm in a goroutine
	go func() {
		cancelChan := make(chan os.Signal, 1)
		// catch SIGTERM or SIGINT
		signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-cancelChan
		log.Warn().Str("signal", sig.String()).Msg("Received signal")

		// Callback to the service
		err := onTerminate(sig)
		if err != nil {
			log.Err(err).Msg("Error during termination")
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()

	// Fetch and parse service definition as injected by roverd
	definition := os.Getenv("ASE_SERVICE")
	if definition == "" {
		log.Fatal().Msg("No service definition found in environment variable ASE_SERVICE. Are you sure that this service is started by roverd?")
	}

	service, err := UnmarshalInjectedService([]byte(definition))
	if err != nil {
		log.Fatal().Err(err).Msg("Error unmarshalling service definition")
	}

	// Enable logging using zerolog
	setupLogging(*debug, *output, service)

	// Create a configuration for this service that will be shared with the user program
	configuration := NewServiceConfiguration(service)

	// Support ota tuning in this goroutine
	// (the user program can fetch the latest value from the configuration)
	if *service.Tuning.Enabled {
		go func() {
			// Initialize zmq socket to retrieve OTA tuning values from the service responsible for this
			socket, err := zmq4.NewSocket(zmq4.REQ)
			if err != nil {
				log.Err(err).Msg("Failed to create socket for OTA tuning")
				return
			}
			defer socket.Close()

			err = socket.Connect(*service.Tuning.Address)
			if err != nil {
				log.Err(err).Msg("Failed to connect to OTA tuning service")
				return
			}
			for {
				// Receive new configuration, and update this in the shared configuration
				_, err := socket.Recv(0) // _ = res
				if err != nil {
					log.Err(err).Msg("Failed to receive tuning values")
					continue
				}

				// todo: use RES!
			}
		}()
	}

	// Run the user program
	err = main(
		service,
		configuration,
	)

	// Handle termination
	if err != nil {
		log.Err(err).Msg("Service quit unexpectedly. Exiting...")
		os.Exit(1)
	} else {
		log.Info().Msg("Service finished successfully")
	}
}
