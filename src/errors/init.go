package customerrors

import "fmt"

//
// Errors that are used within the ServiceRunner
// To be checked using errors.Is() and errors.As()
//

// Error returned when a dependency is not exposed by the service
var OutputNotExposed = fmt.Errorf("Service does not expose requested output")
var ServiceNotRunning = fmt.Errorf("Service is not running (yet)")
