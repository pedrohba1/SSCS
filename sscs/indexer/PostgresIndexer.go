package indexer

import (
	BaseLogger "sscs/logger"
	"sscs/recorder"
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RecordedEvent = recorder.RecordedEvent

type PostgresIndexer struct {
	dsn    string
	db     *gorm.DB
	logger *logrus.Entry

	wg        sync.WaitGroup
	recordOut <-chan RecordedEvent
	stopCh    chan struct{}
}

func NewEventIndexer(dsn string, recordOut <-chan RecordedEvent) (*PostgresIndexer, error) {
	p := &PostgresIndexer{dsn: dsn, recordOut: recordOut, stopCh: make(chan struct{})}
	p.setupLogger()

	return p, nil
}

func (p *PostgresIndexer) Stop() error {
	p.logger.Info("shutting server")
	close(p.stopCh)
	p.wg.Wait() // Wait for the recording goroutine to finish
	return nil
}

func (p *PostgresIndexer) AutoMigrate() error {
	p.logger.Info("migrating tables...")
	return p.db.AutoMigrate(&recorder.RecordedEvent{})
}

func (p *PostgresIndexer) SaveRecord(event RecordedEvent) error {

	err := p.db.Create(&event).Error
	if err != nil {
		p.logger.Info("error indexing record")
	}
	p.logger.Info("saved record: %w", event)
	return err
}

func (p *PostgresIndexer) setupLogger() {
	p.logger = BaseLogger.BaseLogger.WithField("package", "indexer")
}

// TODO: mover conexão com banco para o start para poder dar o stop sem matar o objeto
func (p *PostgresIndexer) Start() error {

	p.logger.Infof("connecting postgres...")
	db, err := gorm.Open(postgres.Open(p.dsn), &gorm.Config{})
	if err != nil {
		p.logger.Error("failed to parse url: %w", err)
		return err
	}

	p.db = db
	p.AutoMigrate()
	p.wg.Add(1)
	go p.listen()

	return nil
}

func (p *PostgresIndexer) listen() error {
	defer p.wg.Done()

	p.logger.Info("listening to index events...")

	for {
		select {
		case <-p.stopCh:
			p.logger.Info("Received stop signal, stopping indexer.")
			return nil // stop signal received, so we return from the function
		case record := <-p.recordOut:
			// New record received, we should save it
			if err := p.SaveRecord(record); err != nil {
				p.logger.Errorf("Failed to save record: %v", err)
			}
		}
	}

}
