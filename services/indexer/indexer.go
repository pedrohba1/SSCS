package indexer

// an interface for indexing data into a database
type Indexer interface {
	Start() error
	Stop() error
	setupLogger()
	listen() error
}
