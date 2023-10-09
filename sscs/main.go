// Example of a daemon with echo service
package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"sscs/core"
	BaseLogger "sscs/logger"

	"github.com/sirupsen/logrus"
	"github.com/takama/daemon"
)

const (

	// name of the service
	name        = "scss"
	description = "Self-sovereign camera system"

	// port which daemon should be listen
	port = ":9977"
)

// dependencies that are NOT required by the service, but might be used.
// in our case, our daemon relies on an available network connection.
// so it is advised to have your network service as a dependency
var dependencies = []string{"NetworkManager.service"}

var logger *logrus.Entry

// Service has embedded daemon
type Service struct {
	daemon.Daemon
	core *core.Core
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: myservice install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	// Do something, call your goroutines, etc

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Set up listener for defined host and port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return "Possibly was a problem with the port binding", err
	}

	// Initialize the Core application
	args := []string{""}
	service.core = core.New(args)
	service.core.Logger.Info("I'm completely operational, and all my circuits are functioning perfectly")
	// loop work cycle with accept connections or interrupt
	// by system signal
	for {
		select {
		case killSignal := <-interrupt:
			logger.Info("Got signal:", killSignal)
			logger.Info("Stoping listening on ", listener.Addr())
			listener.Close()
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}

	// never happen, but need to complete code
	return usage, nil
}

func init() {
	logger = BaseLogger.BaseLogger.WithField("package", "main")
}

func main() {
	srv, err := daemon.New(name, description, daemon.SystemDaemon, dependencies...)

	if err != nil {
		logger.Error("Error: ", err)
		os.Exit(1)
	}
	service := &Service{Daemon: srv}
	status, err := service.Manage()
	if err != nil {
		logger.Error(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)

}
