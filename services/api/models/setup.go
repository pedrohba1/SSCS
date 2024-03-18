// models/setup.go

package models

import (
	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/recognizer"
	"github.com/pedrohba1/SSCS/services/recorder"
	"github.com/pedrohba1/SSCS/services/storer"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	cfg, _ := conf.ReadConf()

	dsn := cfg.Indexer.DbUrl

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("error connecting to database")
	}
	db.AutoMigrate(&recorder.RecordedEvent{})
	db.AutoMigrate(&recognizer.RecognizedEvent{})
	db.AutoMigrate(&storer.CleanedEvent{})
	DB = db
}
