package indexer

import (
	BaseLogger "sscs/logger"
	"sscs/recorder"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RecordedEvent = recorder.RecordedEvent

type PostgresIndexer struct {
	db     *gorm.DB
	logger *logrus.Entry
}

func NewEventIndexer(dsn string) (*PostgresIndexer, error) {
	p := &PostgresIndexer{}
	p.setupLogger()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		p.logger.Error("failed to parse url: %w", err)
		return nil, err
	}

	p.db = db
	return p, nil
}

// TODO: mover conexão com banco para o start para poder dar o stop sem matar o objeto
func (indexer *PostgresIndexer) Start() error {
	indexer.logger.Infof("starting server")
	return nil
}

func (indexer *PostgresIndexer) Stop() error {
	indexer.logger.Infof("shutting server")
	return nil
}
func (indexer *PostgresIndexer) AutoMigrate() error {
	return indexer.db.AutoMigrate(&recorder.RecordedEvent{})
}

func (indexer *PostgresIndexer) SaveEvent(event RecordedEvent) error {
	return indexer.db.Create(&event).Error
}

func (p *PostgresIndexer) setupLogger() {
	p.logger = BaseLogger.BaseLogger.WithField("package", "recorder")
}

// a connection example here
// func main() {
// 	// Your PostgreSQL connection string.
// 	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
// 	indexer, err := NewEventIndexer(dsn)
// 	if err != nil {
// 		log.Fatalf("Failed to create indexer: %s", err)
// 	}

// 	// Automatically migrate your schema, to keep your schema update-to-date.
// 	if err := indexer.AutoMigrate(); err != nil {
// 		log.Fatalf("Failed to auto-migrate: %s", err)
// 	}

// 	event := RecordedEvent{
// 		VideoName: "sample_video.mp4",
// 		Timestamp: time.Now(),
// 	}

// 	if err := indexer.SaveEvent(event); err != nil {
// 		log.Fatalf("Failed to save event: %s", err)
// 	}

// 	fmt.Println("Event saved successfully!")
// }
