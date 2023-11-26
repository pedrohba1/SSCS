// Package recorgnizer provides implementations
// for multiple image recognition algorithms.
package recorgnizer

type Recorgnizer interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
	setupLogger()
	view() error
}
