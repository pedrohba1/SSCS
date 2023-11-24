package core

import (
	"context"
	"image"
	"os"
	"os/signal"
	"sscs/conf"
	"sscs/indexer"
	BaseLogger "sscs/logger"
	"sscs/recorder"
	"sscs/recorgnizer"

	"github.com/sirupsen/logrus"
)

type Core struct {
	ctx         context.Context
	ctxCancel   func()
	configPath  string
	config      conf.Config
	Logger      *logrus.Entry
	recorder    recorder.Recorder
	indexer     indexer.Indexer
	recorgnizer recorgnizer.Recorgnizer
	done        chan struct{}
}

// creates a new core
func New(args []string) *Core {
	configPath := args[0] // assumes args[0] has the config path

	// Read the configuration using ReadConf
	cfg, err := conf.ReadConf()
	if err != nil {
		panic(err) // or handle the error more gracefully, based on your application's needs
	}

	// start resources
	recordChan := make(chan recorder.RecordedEvent, 1)
	frameChan := make(chan image.Image, 1)
	r := recorder.NewRTSP_H264Recorder(cfg.Recorder.RTSP.Feeds[0], recordChan, frameChan)
	r.Start()

	dsn := cfg.Indexer.DbUrl
	i, err := indexer.NewEventIndexer(dsn, recordChan)

	if err != nil {
		panic(err) // or handle the error more gracefully, based on your application's needs
	}

	i.Start()
	v := recorgnizer.NewFaceDetector(frameChan)
	v.Start()

	ctx, ctxCancel := context.WithCancel(context.Background())

	// Create a new Core instance with the read configuration
	p := &Core{
		ctx:         ctx,
		ctxCancel:   ctxCancel,
		configPath:  configPath,
		config:      cfg,
		recorder:    r,
		indexer:     i,
		recorgnizer: v,
		Logger:      BaseLogger.BaseLogger.WithField("package", "core"),
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
	if p.recorgnizer != nil {
		p.recorgnizer.Stop()
	}
	if p.indexer != nil {
		p.indexer.Stop()
	}
	if p.recorder != nil {
		p.recorder.Stop()
	}
}
