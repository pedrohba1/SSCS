// Package cleaner contains the cleaner component implementations.
//
// It handles the storage of recordings, so it doesn't extrapolate
// a limit. It can do that by either deleting the oldest files or backing them
// up on some external storage.
package cleaner

// Cleaner is an interface for a cleaner component
//
// As per definition of SSCS components, a cleaner
// must implement the Start(), Stop() and setupLogger() methods.
type Cleaner interface {
	Start() error
	Stop() error
	setupLogger()
	listen() error
}

// CleantEvent is useful to emit events to
// other components (such as the indexer)
// after deletion of some file
type CleanEvent struct {
	filename string
	fileSize int
}
