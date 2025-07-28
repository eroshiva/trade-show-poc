// Package manager_test implements unit tests to test the control loop behavior.
package manager_test

import (
	"os"
	"testing"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/pkg/connectors"
	simulatorv1 "github.com/eroshiva/trade-show-poc/pkg/mocks"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
)

const (
	host1     = "localhost"
	port1     = "50151"
	protocol1 = endpoint.ProtocolPROTOCOL_NETCONF

	host2     = "localhost"
	port2     = "50152"
	protocol2 = endpoint.ProtocolPROTOCOL_SNMP

	host3     = "localhost"
	port3     = "50153"
	protocol3 = endpoint.ProtocolPROTOCOL_RESTCONF

	host4     = "localhost"
	port4     = "50154"
	protocol4 = endpoint.ProtocolPROTOCOL_OPEN_V_SWITCH

	testControlLoopPeriod = 250 * time.Millisecond

	defaultGRPCTestServerAddress = "localhost:50251"
	defaultHTTPTestServerAddress = "localhost:50252"
)

var (
	client     *ent.Client
	grpcClient apiv1.DeviceMonitoringServiceClient
)

func TestMain(m *testing.M) {
	var err error
	entClient, serverClient, wg, termChan, reverseProxyTermChan, err := monitoring_testing.SetupFull(defaultGRPCTestServerAddress, defaultHTTPTestServerAddress)
	if err != nil {
		panic(err)
	}
	client = entClient
	grpcClient = serverClient

	// creating endpoints and device simulators

	// running tests
	code := m.Run()

	// all tests were run, stopping servers gracefully
	monitoring_testing.TeardownFull(client, wg, termChan, reverseProxyTermChan)
	os.Exit(code)
}

func TestMainControlLoop(t *testing.T) {
	ds1 := simulatorv1.NewDeviceSimulator()
	ds2 := simulatorv1.NewDeviceSimulator()
	ds3 := simulatorv1.NewDeviceSimulator()
	ds4 := simulatorv1.NewDeviceSimulator()

	// starting first simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host1, port1))
	ds1.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds1.StopNetworkDeviceSimulator()
	})

	// starting second simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host2, port2))
	ds2.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds2.StopNetworkDeviceSimulator()
	})

	// starting third simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host3, port3))
	ds3.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds3.StopNetworkDeviceSimulator()
	})

	// starting second simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host4, port4))
	ds4.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds4.StopNetworkDeviceSimulator()
	})

	time.Sleep(testControlLoopPeriod)

	// initial setup is complete, now starting the main control loop
}
