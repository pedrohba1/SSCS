package indexer

// Recorder is an interface for recording streams.
type Indexer interface {
	Start() error // Starts the recording
	Stop() error  // Stops the recording
	setupLogger()
}
