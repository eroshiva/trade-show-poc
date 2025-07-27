// Package server_test implements unit tests for gRPC server functions.
package server_test

import (
	"os"
	"testing"

	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
)

func TestMain(m *testing.M) {
	var err error
	entClient, wg, termChan, reverseProxyTermChan, err := monitoring_testing.SetupFull()
	if err != nil {
		panic(err)
	}
	client = entClient

	// running tests
	code := m.Run()

	// all tests were run, stopping servers gracefully
	monitoring_testing.TeardownFull(client, wg, termChan, reverseProxyTermChan)
	os.Exit(code)
}

func TestAddDevice(t *testing.T) {
}
