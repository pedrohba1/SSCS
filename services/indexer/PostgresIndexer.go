package indexer

import (
	"sync"

	BaseLogger "github.com/pedrohba1/SSCS/services/logger"
	"github.com/pedrohba1/SSCS/services/recognizer"
	"github.com/pedrohba1/SSCS/services/recorder"
	"github.com/pedrohba1/SSCS/services/storer"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresIndexer struct {
	dsn    string
	db     *gorm.DB
	logger *logrus.Entry

	wg     sync.WaitGroup
	eChans EventChannels
	stopCh chan struct{}
}

func NewEventIndexer(dsn string, eChans EventChannels) (*PostgresIndexer, error) {
	p := &PostgresIndexer{dsn: dsn,
		eChans: eChans,
		stopCh: make(chan struct{})}
	p.setupLogger()

	return p, nil
}

func (p *PostgresIndexer) Stop() error {
	close(p.stopCh)
	p.wg.Wait() // Wait for the recording goroutine to finish
	return nil
}

func (p *PostgresIndexer) AutoMigrate() error {
	p.logger.Info("migrating tables...")
	p.db.AutoMigrate(&recorder.RecordedEvent{})
	p.db.AutoMigrate(&recognizer.RecognizedEvent{})
	p.db.AutoMigrate(&storer.CleanedEvent{})
	return nil
}

func (p *PostgresIndexer) saveRecord(event recorder.RecordedEvent) error {
	err := p.db.Create(&event).Error
	if err != nil {
		p.logger.Info("error indexing record")
	}
	p.logger.Info("saved record: %w", event)
	return err
}

func (p *PostgresIndexer) saveRecognition(event recognizer.RecognizedEvent) error {
	err := p.db.Create(&event).Error
	if err != nil {
		p.logger.Info("error indexing record")
	}
	p.logger.Info("saved record: %w", event)
	return err
}

func (p *PostgresIndexer) modifyCleaned(event storer.CleanedEvent) error {
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
			p.logger.Info("Received stop signal")
			return nil
		case record := <-p.eChans.RecordOut:
			if err := p.saveRecord(record); err != nil {
				p.logger.Errorf("Failed to save record: %v", err)
			}

		case recog := <-p.eChans.RecogOut:
			if err := p.saveRecognition(recog); err != nil {
				p.logger.Errorf("Failed to save record: %v", err)
			}
		case clean := <-p.eChans.CleanOut:
			if err := p.modifyCleaned(clean); err != nil {
				p.logger.Errorf("Failed to save record: %v", err)
			}
		}
	}

}
