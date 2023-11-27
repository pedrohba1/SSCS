package core

import (
	"context"
	"image"
	"os"
	"os/signal"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/indexer"
	BaseLogger "github.com/pedrohba1/SSCS/services/logger"
	"github.com/pedrohba1/SSCS/services/recognizer"
	"github.com/pedrohba1/SSCS/services/recorder"
	"github.com/pedrohba1/SSCS/services/storer"

	"github.com/sirupsen/logrus"
)

type Core struct {
	ctx        context.Context
	ctxCancel  func()
	configPath string
	config     conf.Config
	Logger     *logrus.Entry
	recorder   recorder.Recorder
	indexer    indexer.Indexer
	storer     storer.Storer
	recognizer recognizer.Recognizer
	done       chan struct{}
}

// creates a new core
func New(args []string) *Core {
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
	r.Start()

	// starts the recognizer
	recogChan := make(chan recognizer.RecognizedEvent, 5)
	v := recognizer.NewFaceDetector(recognizer.EventChannels{
		FrameIn: frameChan,
	})
	v.Start()

	ctx, ctxCancel := context.WithCancel(context.Background())

	// starts the cleaner
	cleanChan := make(chan storer.CleanedEvent)
	s := storer.NewOSStorer(cleanChan)
	s.Start()

	// starts the indexer
	dsn := cfg.Indexer.DbUrl
	i, err := indexer.NewEventIndexer(dsn, indexer.EventChannels{
		RecordOut: recordChan,
		RecogOut:  recogChan,
		CleanOut:  cleanChan,
	})

	if err != nil {
		panic(err)
	}
	i.Start()

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

	go p.run()

	return p
}

// Close closes Core and waits for all goroutines to return.
func (p *Core) Close() {
	p.closeResources()
	p.ctxCancel()
	<-p.done
}

// Wait waits for the Core to exit.
func (p *Core) Wait() {
	<-p.done
}

func (p *Core) run() {
	defer close(p.done)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

outer:
	for {
		select {
		case <-interrupt:
			p.Logger.Info("shutting down gracefully")
			break outer

		case <-p.ctx.Done():
			break outer
		}
	}

}

func (p *Core) closeResources() {
	if p.recognizer != nil {
		p.recognizer.Stop()
	}
	if p.indexer != nil {
		p.indexer.Stop()
	}
	if p.recorder != nil {
		p.recorder.Stop()
	}
	if p.storer != nil {
		p.storer.Stop()
	}
}
