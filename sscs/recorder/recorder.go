package recorder

// Recorder is an interface for recording streams.
type Recorder interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
}
