// Package indexer provides implementations
// for database technologies that integrate with SSCS
package indexer

import (
	"github.com/pedrohba1/SSCS/services/recognizer"
	"github.com/pedrohba1/SSCS/services/recorder"
	"github.com/pedrohba1/SSCS/services/storer"
)

// Indexer is an interface for an indexer component.
//
// As per definition of SSCS components, an indexer
// must implement the Start(), Stop() and setupLogger() methods.
type Indexer interface {
	Start() error
	Stop() error
	setupLogger()
	listen() error
}

// Available channels for communicating with this service.
//
// Notice that they are all channels for receiving
// data. The indexer is only supposed to receive information
// from these sources.
type EventChannels struct {
	RecordOut <-chan recorder.RecordedEvent
	RecogOut  <-chan recognizer.RecognizedEvent
	CleanOut  <-chan storer.CleanedEvent
}
