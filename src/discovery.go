package servicerunner

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	pb_core_messages "github.com/VU-ASE/rovercom/packages/go/core"
	customerrors "github.com/VU-ASE/roverlib/src/errors"
	zmq "github.com/pebbe/zmq4"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

// Checks if all depdnencies are resolved by checking if the list of resolved dependencies contains each dependency
func allDependenciesResolved(service serviceDefinition, resolvedDependencies []ResolvedDependency) bool {
	if len(service.Dependencies) == 0 {
		return true
	}

	for _, dependency := range service.Dependencies {
		found := false
		for _, resolvedDependency := range resolvedDependencies {
			if strings.EqualFold(dependency.ServiceName, resolvedDependency.ServiceName) && strings.EqualFold(dependency.OutputName, resolvedDependency.OutputName) {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// Will contact the discovery service to get the addresses of each dependency and register this service with the service discovery service (the system manager)
func registerService(service serviceDefinition, sysmanReqRepAddr string) ([]ResolvedDependency, error) {
	// create a zmq client socket to the system manager
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("Could not open ZMQ connection to system manager: %s", err)
	}
	defer client.Close()
	log.Debug().Str("service", service.Name).Str("address", sysmanReqRepAddr).Msg("Connecting to system manager")
	err = client.Connect(sysmanReqRepAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to system manager: %s", err)
	}

	// convert our service definition to a protobuf message
	endpoints := []*pb_core_messages.ServiceEndpoint{}
	for _, output := range service.Outputs {
		// convert our struct to the ServiceEndpoint protobuf message
		endpoints = append(endpoints, &pb_core_messages.ServiceEndpoint{
			Name:    output.Name,
			Address: output.Address,
		})
	}
	options := []*pb_core_messages.ServiceOption{}
	for _, option := range service.Options {
		// convert our struct to the ServiceOption protobuf message
		newOption := &pb_core_messages.ServiceOption{
			Name:    option.Name,
			Mutable: option.Mutable,
		}
		if option.DefaultValue != "" {
			switch option.Type {
			case "string":
				newOption.Type = pb_core_messages.ServiceOption_STRING
				newOption.StringDefault = option.DefaultValue
			case "int":
				newOption.Type = pb_core_messages.ServiceOption_INT
				intval, err := strconv.Atoi(option.DefaultValue)
				if err != nil {
					return nil, fmt.Errorf("Option '%s' has type int, but a default value that is not an int: %s", option.Name, option.DefaultValue)
				} else {
					newOption.IntDefault = int32(intval)
				}
			case "float":
				newOption.Type = pb_core_messages.ServiceOption_FLOAT
				floatval, err := strconv.ParseFloat(option.DefaultValue, 64)
				if err != nil {
					return nil, fmt.Errorf("Option '%s' has type float, but a default value that is not a float: %s", option.Name, option.DefaultValue)
				} else {
					newOption.FloatDefault = float32(floatval)
				}
			default:
				return nil, fmt.Errorf("Option '%s' has an unknown type: %s", option.Name, option.Type)
			}
		}
		options = append(options, newOption)
	}
	dependencies := []*pb_core_messages.ServiceDependency{}
	for _, dependency := range service.Dependencies {
		// convert our struct to the ServiceDependency protobuf message
		dependencies = append(dependencies, &pb_core_messages.ServiceDependency{
			ServiceName: dependency.ServiceName,
			OutputName:  dependency.OutputName,
		})
	}
	// create a registration message
	regMsg := pb_core_messages.CoreMessage{
		Msg: &pb_core_messages.CoreMessage_Service{
			Service: &pb_core_messages.Service{
				Identifier: &pb_core_messages.ServiceIdentifier{
					Name: service.Name,
					Pid:  int32(os.Getpid()),
				},
				Endpoints:    endpoints,
				Options:      options,
				Dependencies: dependencies,
			},
		},
	}

	// convert the message to bytes
	msgBytes, err := proto.Marshal(&regMsg)
	if err != nil {
		log.Err(err).Msg("Error marshalling protobuf message")
		return nil, err
	}

	// send registration to the system manager
	log.Info().Str("service", service.Name).Msg("Registering service with system manager")
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
				log.Warn().Str("service", service.Name).Msgf("Still waiting for response from system manager. Are you sure the system manager is running and available at '%s'?", sysmanReqRepAddr)
			} else {
				log.Info().Str("service", service.Name).Msg("Waiting for response from system manager")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// wait for a response from the system manager
	resBytes, err := client.RecvBytes(0)
	responseReceived = true

	// the response must be of type Service (see include/servicediscovery.proto)
	// if not, we discard it: registration not successful
	log.Info().Str("service", service.Name).Msg("Received registration response from system manager")
	if err != nil {
		return nil, err
	}
	response := pb_core_messages.CoreMessage{}
	err = proto.Unmarshal(resBytes, &response)
	if err != nil {
		log.Err(err).Msg("Error unmarshalling protobuf message")
		return nil, err
	}
	errorMessage := response.GetError()
	if errorMessage != nil {
		return nil, fmt.Errorf("System manager denied service registration: %s", errorMessage.Message)
	}
	responseService := response.GetService()
	if responseService == nil {
		return nil, fmt.Errorf("Received empty response from system manager")
	}
	// check if the name and pid of the response match our registration, if not someone else registered with the same name
	identifier := responseService.GetIdentifier()
	if identifier == nil {
		return nil, fmt.Errorf("Received empty response from system manager")
	}
	name := identifier.GetName()
	pid := identifier.GetPid()
	if name != service.Name {
		return nil, fmt.Errorf("System manager denied service registration, name mismatch (registered as %s, but received %s)", service.Name, name)
	}
	if pid != int32(os.Getpid()) {
		return nil, fmt.Errorf("System manager denied service registration, service %s was already registered by pid %d", service.Name, pid)
	}
	// check if the endpoints match our registration
	registeredEndpints := responseService.GetEndpoints()
	for i, endpoint := range endpoints {
		registeredEndpoint := registeredEndpints[i]
		if registeredEndpoint == nil {
			return nil, fmt.Errorf("Endpoint %s was not registered", endpoint.Name)
		} else if registeredEndpoint.GetName() != endpoint.Name || registeredEndpoint.GetAddress() != endpoint.Address {
			return nil, fmt.Errorf("Endpoint %s was registered with different address (%s) than requested (%s)", endpoint.Name, registeredEndpoint.GetAddress(), endpoint.Address)
		}
	}

	// registration was successfull!
	log.Info().Str("service", service.Name).Msg("Service registration successful")

	// Resolve dependencies, always request the system manager broadcast address
	service.Dependencies = append(service.Dependencies, dependency{
		ServiceName: "systemmanager",
		OutputName:  "broadcast",
	})
	return resolveDependencies(service, client)
}

func resolveDependencies(service serviceDefinition, serverSocket *zmq.Socket) ([]ResolvedDependency, error) {
	resolvedDependencies := make([]ResolvedDependency, 0)
	// do we even have dependencies?
	if len(service.Dependencies) <= 0 {
		log.Info().Msg("No dependencies to resolve")
		return resolvedDependencies, nil
	}

	// Now attempt to resolve dependencies, if any
	log.Info().Str("service", service.Name).Int("dependencies to resolve", len(service.Dependencies)).Msg("Resolving dependencies")
	for !allDependenciesResolved(service, resolvedDependencies) {
		// create a list all unique *services* (not endpoints) that we depend on and that are not yet resolved
		uniqueServiceDependencies := extractUniqueServices(service.Dependencies, resolvedDependencies)

		// resolve each service (sequentially)
		for _, serviceName := range uniqueServiceDependencies {
			dependencyInfo, err := requestServiceInformation(serviceName, serverSocket)
			if err != nil {
				if errors.Is(err, customerrors.ServiceNotRunning) {
					log.Warn().Str("dependency", serviceName).Msg("Dependency is not running (yet), will retry in 3 seconds")
					time.Sleep(3 * time.Second)
					continue
				} else {
					log.Err(err).Str("dependency", serviceName).Msg("Error resolving dependency")
					return nil, err
				}
			}

			// fill the list of resolved dependencies using the dependency information
			for _, dependency := range service.Dependencies {
				// this is not optimal, because we iterate over the dependencies and the resolved dependencies for each dependency
				// but the list of dependencies is small, so it should be fine
				// an optimization could be to remove the resolved dependencies from the list of dependencies
				if !dependencyResolved(dependency, resolvedDependencies) && isDependencyOfService(dependency, serviceName) {
					resolvedDependency, err := getDependencyFromServiceInformation(dependencyInfo, dependency)
					if err != nil && errors.Is(err, customerrors.OutputNotExposed) {
						log.Error().Str("dependency", dependency.ServiceName).Str("output", dependency.OutputName).Msgf("Dependency does not expose requested output. Retrying would not help, since the output definition will probably not change during runtime. Please update the service definition of service '%s' to make sure to expose '%s'", dependency.ServiceName, dependency.OutputName)
						return nil, err
					} else if err != nil {
						log.Error().Str("dependency", dependency.ServiceName).Str("output", dependency.OutputName).Msg("Error resolving dependency")
						return nil, err
					}

					log.Info().Str("dependency", dependency.ServiceName).Str("output", dependency.OutputName).Msg("Resolved dependency")
					resolvedDependencies = append(resolvedDependencies, resolvedDependency)
				}
			}

		}
	}
	return resolvedDependencies, nil
}

// Extract the unique service names of all unresolved dependencies
func extractUniqueServices(dependencies []dependency, resolvedDependencies []ResolvedDependency) []string {
	uniqueServices := make([]string, 0)
	for _, dependency := range dependencies {
		if !slices.Contains(uniqueServices, dependency.ServiceName) && !slices.ContainsFunc(resolvedDependencies, func(dep ResolvedDependency) bool {
			return strings.EqualFold(dep.ServiceName, dependency.ServiceName) && strings.EqualFold(dep.OutputName, dependency.OutputName)
		}) {
			uniqueServices = append(uniqueServices, dependency.ServiceName)
		}
	}
	return uniqueServices
}

func requestServiceInformation(serviceName string, serverSocket *zmq.Socket) (*pb_core_messages.Service, error) {
	// create a request message
	reqMsg := pb_core_messages.CoreMessage{
		Msg: &pb_core_messages.CoreMessage_ServiceInformationRequest{
			ServiceInformationRequest: &pb_core_messages.ServiceInformationRequest{
				Requested: &pb_core_messages.ServiceIdentifier{
					Name: serviceName,
					Pid:  1, // does not matter
				},
			},
		},
	}

	// convert the message to bytes
	msgBytes, err := proto.Marshal(&reqMsg)
	if err != nil {
		log.Err(err).Msg("Error marshalling protobuf message")
		return nil, err
	}
	// send the request to the system manager
	_, err = serverSocket.SendBytes(msgBytes, 0)
	if err != nil {
		return nil, err
	}

	log.Info().Str("dependency", serviceName).Msg("Requesting dependency information from system manager")
	gotReply := false
	go func() {
		count := 0
		for {
			// print a idle message every 5 seconds, until were done
			if gotReply {
				return
			}
			if count > 5 {
				log.Warn().Str("dependency", serviceName).Msg("Still waiting for dependency response from system manager. Are you sure the system manager is running?")
			} else {
				log.Info().Str("dependency", serviceName).Msg("Waiting for dependency response from system manager")
			}
			time.Sleep(5 * time.Second)
			count++
		}
	}()
	// wait for a response from the system manager
	resBytes, err := serverSocket.RecvBytes(0)
	gotReply = true
	if err != nil {
		return nil, err
	}

	// parse the response
	// the response must be of type Service (see messages/servicediscovery.proto)
	response := pb_core_messages.CoreMessage{}
	err = proto.Unmarshal(resBytes, &response)
	respondedService := response.GetService()
	if respondedService == nil {
		return nil, fmt.Errorf("Received empty response from system manager, expected Service")
	}
	if err != nil {
		return nil, err
	} else if respondedService.Status != pb_core_messages.ServiceStatus_RUNNING {
		// pass a detectable error, so that the caller can retry later
		return nil, customerrors.ServiceNotRunning
	}
	// service is running!
	return respondedService, nil
}

// Check if a dependency is already resolved (by checking if it is in the list of resolved dependencies)
func dependencyResolved(dependency dependency, resolvedDependencies []ResolvedDependency) bool {
	for _, resolvedDependency := range resolvedDependencies {
		if dependency.ServiceName == resolvedDependency.ServiceName && dependency.OutputName == resolvedDependency.OutputName {
			return true
		}
	}
	return false
}

// Used to filter out dependencies that cannot be resolved by this service information
// e.g. the dependency serviceB.outputA cannot be resolved by serviceC, but it can be resolved by serviceA
func isDependencyOfService(dependency dependency, serviceName string) bool {
	return strings.EqualFold(dependency.ServiceName, serviceName)
}

// Returns a resolved dependency, given a service status and a dependency
func getDependencyFromServiceInformation(service *pb_core_messages.Service, dependency dependency) (ResolvedDependency, error) {
	if service == nil {
		return ResolvedDependency{}, fmt.Errorf("Received empty service status")
	}

	// check if the service exposes the output that we need
	endpoints := service.GetEndpoints()
	if endpoints == nil {
		return ResolvedDependency{}, fmt.Errorf("Received empty service endpoints")
	}

	for _, endpoint := range endpoints {
		if strings.EqualFold(endpoint.GetName(), dependency.OutputName) {
			return ResolvedDependency{
				ServiceName: dependency.ServiceName,
				OutputName:  dependency.OutputName,
				Address:     endpoint.GetAddress(),
			}, nil
		}
	}

	return ResolvedDependency{}, customerrors.OutputNotExposed
}

// Get a list of all services
func getServiceList(sysmanReqRepAddr string) (*pb_core_messages.ServiceList, error) {
	// Create a zmq client socket to the system manager
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("Could not open ZMQ connection to system manager: %s", err)
	}
	defer client.Close()
	log.Debug().Str("address", sysmanReqRepAddr).Msg("Connecting to system manager")
	err = client.Connect(sysmanReqRepAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to system manager: %s", err)
	}

	// Create a request message
	reqMsg := pb_core_messages.CoreMessage{
		Msg: &pb_core_messages.CoreMessage_ServiceListRequest{
			ServiceListRequest: &pb_core_messages.ServiceListRequest{},
		},
	}
	// Marshal message
	msgBytes, err := proto.Marshal(&reqMsg)
	if err != nil {
		return nil, err
	}
	// Send the request to the system manager
	_, err = client.SendBytes(msgBytes, 0)
	if err != nil {
		return nil, err
	}
	// Wait for a response from the system manager
	resBytes, err := client.RecvBytes(0)
	if err != nil {
		return nil, err
	}
	// Parse the response
	response := pb_core_messages.CoreMessage{}
	err = proto.Unmarshal(resBytes, &response)
	if err != nil {
		return nil, err
	}

	serviceList := response.GetServiceList()
	if serviceList == nil {
		return nil, fmt.Errorf("Received empty response from system manager")
	}
	return serviceList, nil
}

// Get the tuning state
func getTuningState(sysmanReqRepAddr string) (*pb_core_messages.TuningState, error) {
	// Create a zmq client socket to the system manager
	client, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, fmt.Errorf("Could not open ZMQ connection to system manager: %s", err)
	}
	defer client.Close()
	log.Debug().Str("address", sysmanReqRepAddr).Msg("Connecting to system manager")
	err = client.Connect(sysmanReqRepAddr)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to system manager: %s", err)
	}

	// Create a request message
	reqMsg := pb_core_messages.CoreMessage{
		Msg: &pb_core_messages.CoreMessage_TuningStateRequest{
			TuningStateRequest: &pb_core_messages.TuningStateRequest{},
		},
	}
	// Marshal message
	msgBytes, err := proto.Marshal(&reqMsg)
	if err != nil {
		return nil, err
	}
	// Send the request to the system manager
	_, err = client.SendBytes(msgBytes, 0)
	if err != nil {
		return nil, err
	}
	// Wait for a response from the system manager
	resBytes, err := client.RecvBytes(0)
	if err != nil {
		return nil, err
	}
	// Parse the response
	response := pb_core_messages.CoreMessage{}
	err = proto.Unmarshal(resBytes, &response)
	if err != nil {
		return nil, err
	}

	tuningState := response.GetTuningState()
	if tuningState == nil {
		return nil, fmt.Errorf("Received empty response from system manager")
	}
	return tuningState, nil
}

// Used to update your own service status
func updateServiceStatus(
	sysmanReqRepAddr string,
	identifier *pb_core_messages.ServiceIdentifier,
	status pb_core_messages.ServiceStatus,
) error {
	// create a zmq client socket to the system manager
	socket, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return fmt.Errorf("Could not open ZMQ connection to system manager: %s", err)
	}
	defer socket.Close()
	err = socket.Connect(sysmanReqRepAddr)
	if err != nil {
		return fmt.Errorf("Could not connect to system manager: %s", err)
	}

	// Create a request message
	msg := pb_core_messages.CoreMessage{
		Msg: &pb_core_messages.CoreMessage_ServiceStatusUpdate{
			ServiceStatusUpdate: &pb_core_messages.ServiceStatusUpdate{
				Status:  status,
				Service: identifier,
			},
		},
	}

	// Convert the message to bytes
	msgBytes, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	// Send the request to the system manager
	_, err = socket.SendBytes(msgBytes, 0)
	if err != nil {
		return err
	}

	// We always need to wait for a response, because of the REQ/REP pattern
	_, err = socket.RecvBytes(0)
	if err != nil {
		return err
	}

	// We don't actually care about the response
	return nil
}
