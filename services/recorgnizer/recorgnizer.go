// Recorder is an interface for recording streams.
package recorgnizer

type Recorgnizer interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
	setupLogger()
	view() error
}
