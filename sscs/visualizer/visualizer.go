// Recorder is an interface for recording streams.
package visualizer

type Visualizer interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
	setupLogger()
	view() error
}
