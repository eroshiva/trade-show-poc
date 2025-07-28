// Package main is an entry pint for a device simulator.
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	simulatorv1 "github.com/eroshiva/trade-show-poc/pkg/mocks"
)

func main() {
	termChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	ds := simulatorv1.NewDeviceSimulator()
	ds.StartNetworkDeviceSimulator()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-sigChan
		close(termChan)

		// gracefully close network device simulator server
		ds.StopNetworkDeviceSimulator()
		wg.Done()
	}()

	wg.Wait()
}
