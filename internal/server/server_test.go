// Package server_test implements unit tests for gRPC server functions.
package server_test

import (
	"os"
	"sync"
	"testing"

	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
)

func TestMain(m *testing.M) {
	var err error
	client, err = monitoring_testing.Setup()
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	termChan := make(chan bool, 1)
	readyChan := make(chan bool, 1)
	reverseProxyReadyChan := make(chan bool, 1)
	reverseProxyTermChan := make(chan bool, 1)
	wg.Add(1)
	go func() {
		wg.Add(1) //nolint:staticcheck
		server.StartServer(client, wg, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
		wg.Done()
	}()
	// Waiting until both servers are up and running
	<-readyChan
	<-reverseProxyReadyChan

	code := m.Run()

	// all tests were run, stopping servers gracefully
	close(termChan)
	close(reverseProxyTermChan)
	err = db.GracefullyCloseDBClient(client)
	if err != nil {
		panic(err)
	}
	wg.Wait()

	os.Exit(code)
}

func TestAddDevice(_ *testing.T) {
}
