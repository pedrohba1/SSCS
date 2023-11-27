// Package storer contains the Storer component implementations.
//
// It handles the storage of recordings, so it doesn't extrapolate
// a limit. It can do that by either deleting the oldest files or backing them
// up on some external storage.
package storer

// Storer is an interface for a Storer component
//
// As per definition of SSCS components, a Storer
// must implement the Start(), Stop() and setupLogger() methods.
type Storer interface {
	Start() error
	Stop() error
	setupLogger()
	monitor() error
}

// CleantEvent is useful to emit events to
// other components (such as the indexer)
// after deletion or replacement of some file
type CleanEvent struct {
	filename string
	fileSize int
}
