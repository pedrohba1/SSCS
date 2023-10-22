package recorder

import "time"

// Recorder is an interface for recording streams.
type Recorder interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
	SetupLogger()
}

type RecordedEvent struct {
	VideoName string
	Timestamp time.Time
}
