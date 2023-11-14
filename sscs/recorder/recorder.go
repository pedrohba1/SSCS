package recorder

import (
	"image"
	"sync"
	"time"
)

// Recorder is an interface for recording streams.
type Recorder interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
	setupLogger()
	record() error
}

type RecordedEvent struct {
	VideoName string
	StartTime time.Time
	EndTime   time.Time
}

type CameraFeed struct {
	sync.RWMutex
	Frame image.Image
}
