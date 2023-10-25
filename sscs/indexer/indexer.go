package indexer

type Indexer interface {
	Start() error
	Stop() error
	setupLogger()
}
