// Package core provides the central coordination logic for a multimedia processing system.
// It integrates several services such as recording, indexing, storing, and recognizing,
// and handles graceful shutdowns and resource management.
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

// Core coordinates the various components of the system. It manages the lifecycle of
// services like the recorder, indexer, storer, and recognizer, ensuring that they start,
// operate, and stop gracefully. Core also handles application-level configurations and logging.
type Core struct {
	ctx        context.Context    // Application context used to manage cancellation.
	ctxCancel  func()             // Function to call to cancel the application context.
	configPath string             // Path to the configuration file.
	config     conf.Config        // Struct holding the loaded configuration.
	Logger     *logrus.Entry      // Logger instance for logging throughout the application.
	recorder   recorder.Recorder  // Component responsible for recording media.
	indexer    indexer.Indexer    // Component responsible for indexing recorded media.
	storer     storer.Storer      	  // Component responsible for storing media.
	recognizer recognizer.Recognizer // Component responsible for recognizing elements in media.
	done       chan struct{}      	// Channel to signal the completion of Core operations.
}


// Close gracefully shuts down all components of Core and waits for all goroutines
// associated with the components to finish. This ensures that all resources are properly
// cleaned up and that there are no resource leaks.
func (p *Core) Close() {
	p.closeResources()
	p.ctxCancel()
	
	<-p.done
}

// Wait blocks until the Core has fully stopped, ensuring that all operations are
// cleanly terminated before the application exits.
func (p *Core) Wait() {
	<-p.done
}


// Start begins the operation of all system components and monitors for interrupt
// signals to initiate a graceful shutdown. This method is the main entry point for
// running the Core's services and should be called after all configurations are set.
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
