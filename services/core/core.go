package core

import (
	"context"
	"os"
	"os/signal"

	"github.com/pedrohba1/SSCS/services/conf"
	"github.com/pedrohba1/SSCS/services/indexer"
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

func (p *Core) Start() {
	defer close(p.done)
	// Start your components
	p.recorder.Start()
	p.recognizer.Start()
	p.storer.Start()
	p.indexer.Start()

	// Handle interrupts or context cancellation
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

	// Now that the components have run, capture the heap profile here
	// f, err := os.Create("heap.prof")
	// if err != nil {
	// 	log.Fatal("could not create memory profile: ", err)
	// }
	// defer f.Close()
	// if err := pprof.WriteHeapProfile(f); err != nil {
	// 	log.Fatal("could not write memory profile: ", err)
	// }
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
