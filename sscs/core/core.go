package core

import (
	"context"
	"sscs/conf"
	BaseLogger "sscs/logger"
	"sscs/recorder"

	"github.com/sirupsen/logrus"
)

type Core struct {
	ctx        context.Context
	ctxCancel  func()
	configPath string
	config     conf.Config
	Logger     *logrus.Entry
	recorder   recorder.Recorder

	// out
	done chan struct{}
}

//creates a new core. The basic functionalities are recording,
func New(args []string) *Core {
	// Extract the config path from the args or define a default
	// This is just a placeholder; you might extract it differently
	configPath := args[0] // assumes args[0] has the config path

	// Read the configuration using ReadConf
	cfg, err := conf.ReadConf("./sscs.yml")
	if err != nil {
		panic(err) // or handle the error more gracefully, based on your application's needs
	}

	// start resources
	r := recorder.NewRTSPRecorder(cfg.RTSP.Feeds[0])
	r.Start()

	ctx, ctxCancel := context.WithCancel(context.Background())

	// Create a new Core instance with the read configuration
	return &Core{
		ctx:        ctx,
		ctxCancel:  ctxCancel,
		configPath: configPath,
		config:     cfg,
		recorder:   r,
		Logger:     BaseLogger.BaseLogger.WithField("package", "core"),
	}
}

// Close closes Core and waits for all goroutines to return.
func (p *Core) Close() {
	p.ctxCancel()
	<-p.done
}

// Wait waits for the Core to exit.
func (p *Core) Wait() {
	<-p.done
}
