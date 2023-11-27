// Package recognizer provides implementations
// for multiple image recognition algorithms.
package recognizer

import "image"

// Recognizer is an interface for a recognizer component.
//
// As per definition of SSCS components, a recognizer
// must implement the Start(), Stop() and setupLogger() methods.
type Recognizer interface {
	Start() error
	Stop() error
	setupLogger()
	view() error
}

// EventChannels are channels for communicating with this service.
type EventChannels struct {
	FrameIn  <-chan image.Image
	RecogOut chan<- RecognizedEvent
}

// Config contains all parameters that can be customized
// via the sscs.yml file.
type Config struct {
	ThumbsDir string
}

// RecognizedEvent is useful to emit events to
// other components (such as the indexer)
// after something was detected by the recognition algorithms
type RecognizedEvent struct {
	eventName string
	timestamp int
}
