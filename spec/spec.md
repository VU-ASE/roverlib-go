# roverlib Go interface and their non-go counterparts
!! all go functions have a second return value (tuple-based) which specifies an error (so `-> int` is actually `-> int, err`)

## start the service, register with core, resolve deps
-- go --
func Run(
    func entry(ts *tuningState), ! the ts object is managed by roverlib and updated on tuning state updates from transceiver, users cannot set properties on it
    func exit,
) -> void

! helper functions that allows non-go wrappers to not need to use any go function pointers
func Register(
    int tuningUpdatesPipe ! the go roverlib will write tuning state updates to this pipe, which the non-go program can read
) -> void
func Deregister() -> void 

--- non-go ---
! this function should execute `Deregister()` after `entry` has returned
func Run(
    func entry(ts *tuningState), ! this object is managed by the roverlib non-go wrapper and updated with values sent over the `Register()` pipe
    func exit ! this is just a SIGINT handler
) -> void

## methods on *tuningState
-- go --
func GetInt/GetFloat/GetString (
    str key
) -> int/float/string

<!-- func GetType (
    str key
) -> string -->

--- non-go ---
func GetInt/GetFloat/GetString (
    str key
) -> int/float/string

<!-- ! But also shorthand function with dynamic type, that uses go's `*ts.GetType()` function internally
func GetValue (
    str key
) -> int | float | string (union) -->

## getting a stream
-- go --
! can be either an input (read) or output (write) stream
! stream should be singleton, no need to reinitialize
func GetStream(
    str name
) -> stream

-- non-go --
func GetStream(
    str name
) -> stream

## methods on a stream
-- go --


! will error if this is a read stream or if the stream is closed
func SendBytes(
    bytes content
    obj{} flags
) -> int ! length send

! only used by non-go libraries to pass flags as ints, used for read as well
func PrimitiveSendBytes(
    bytes content
    int64 flags ! OR'ed flags
)

! will error if this is a read stream or if the stream is closed (based on `SendBytes`)
func Send(
    ! this is a union type of all possible protobuf SensorOutput types
    (CameraOutput | SpeedOutput | RpmOutput | ...) output
) -> int ! length send

! will error if this is a write stream or if the stream is closed
func ReadBytes() -> bytes
! based on `ReadBytes`
func Read() -> (CameraOutput | SpeedOutput | RpmOutput | ...) ! this is a union type of all possible protobuf SensorOutput types







