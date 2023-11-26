// Package recorder contains all the implementations
// for receiving and recording media streams.
// It's implementations support RTSP and H.264 encoding.
package recorder

import (
	"image"
	"time"
)

// Recorder is an interface for recording streams.
type Recorder interface {
	Start() error
	Stop() error
	setupLogger()
	record() error
	sendFrame(image.Image) error
}

type RecordedEvent struct {
	VideoName string
	StartTime time.Time
	EndTime   time.Time
}
