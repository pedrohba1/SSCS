// Package recorder contains all the implementations
// for receiving and recording media streams.
//
// It's implementations support RTSP and H.264 encoding.
package recorder

import (
	"image"
	"time"
)

// Recorder is an interface for a recorder component.
//
// As per definition of SSCS components, a recorder
// must implement the Start(), Stop() and setupLogger() methods.
type Recorder interface {
	Start() error
	Stop() error
	setupLogger()
	record() error
	sendFrame(image.Image) error
}

type EventChannels struct {
	RecordOut chan<- RecordedEvent
	FrameOut  chan<- image.Image
}

// RecordedEvent is used to communicate via channels
// when a recording is saved.
type RecordedEvent struct {
	VideoName string
	StartTime time.Time
	EndTime   time.Time
}
