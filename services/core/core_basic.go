package core

import (
	"context"
	"image"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/indexer"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"
	"github.com/pedrohba1/SSCS/services/recognizer"
	"github.com/pedrohba1/SSCS/services/recorder"
	"github.com/pedrohba1/SSCS/services/storer"
)

// creates a new core, containing the basic setup: a recorder, a recognizer, an indexer and a storer
func NewBasic(args []string) *Core {
	configPath := args[0] // assumes args[0] has the config path

	// Read the configuration using ReadConf
	cfg, err := conf.ReadConf()
	if err != nil {
		panic(err)
	}

	// start resources
	recordChan := make(chan recorder.RecordedEvent, 5)
	frameChan := make(chan image.Image, 10)
	//starts the recorder


	r := recorder.NewRTSP_H264Recorder(cfg.Recorder.RTSP.Feeds[0], recorder.EventChannels{
		RecordOut: recordChan,
		FrameOut:  frameChan,
	})

	// starts the recognizer
	recogChan := make(chan recognizer.RecognizedEvent, 5)
	v := recognizer.NewHaarDetector(recognizer.EventChannels{
		FrameIn: frameChan,
		RecogOut: recogChan,
	})

	ctx, ctxCancel := context.WithCancel(context.Background())

	// starts the cleaner
	cleanChan := make(chan storer.CleanedEvent)
	s := storer.NewOSStorer(storer.EventChannels{CleanOut: cleanChan})

	// starts the indexer
	dsn := cfg.Indexer.DbUrl
	i, err := indexer.NewEventIndexer(dsn, indexer.EventChannels{
		RecordIn: recordChan,
		RecogIn:  recogChan,
		CleanIn:  cleanChan,
	})

	if err != nil {
		panic(err)
	}

	// Create a new Core instance with the read configuration
	p := &Core{
		ctx:        ctx,
		ctxCancel:  ctxCancel,
		configPath: configPath,
		config:     cfg,
		recorder:   r,
		indexer:    i,
		recognizer: v,
		storer:     s,
		Logger:     BaseLogger.BaseLogger.WithField("package", "core"),
	}

	p.done = make(chan struct{})

	go p.Start()

	return p
}

