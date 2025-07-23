// Package main is the main entry point to the service
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/eroshiva/trade-show-poc/internal/server"
)

func main() {
	// channels to handle termination and capture signals
	termChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigChan
		close(termChan)
	}()

	readyChan := make(chan bool, 1)

	server.StartServer(termChan, readyChan)
}
