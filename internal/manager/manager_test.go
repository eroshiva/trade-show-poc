// Package manager_test implements unit tests to test the control loop behavior.
package manager_test

import (
	"context"
	"os"
	"testing"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/manager"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/checksum"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/eroshiva/trade-show-poc/pkg/connectors"
	simulatorv1 "github.com/eroshiva/trade-show-poc/pkg/mocks"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	delta                 = 25 * time.Millisecond

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
	ctx, cancel := context.WithTimeout(context.Background(), 4*monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

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
	time.Sleep(100 * time.Millisecond) // giving some time for the server to bootup

	// starting second simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host2, port2))
	ds2.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds2.StopNetworkDeviceSimulator()
	})
	time.Sleep(100 * time.Millisecond) // giving some time for the server to bootup

	// starting third simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host3, port3))
	ds3.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds3.StopNetworkDeviceSimulator()
	})
	time.Sleep(100 * time.Millisecond) // giving some time for the server to bootup

	// starting second simulator
	t.Setenv(simulatorv1.EnvServerAddress, connectors.CraftServerAddress(host4, port4))
	ds4.StartNetworkDeviceSimulator()
	t.Cleanup(func() {
		ds4.StopNetworkDeviceSimulator()
	})
	time.Sleep(100 * time.Millisecond) // giving some time for the server to bootup

	// initial setup is complete, now starting the main control loop
	// first, creating mock checksum generator
	checksumGen := checksum.NewMockGenerator()
	// creating SB handler
	sbManager := manager.NewManager(client, checksumGen)
	// performing one round of SB handler control loop
	sbManager.PerformControlLoopRoutine(testControlLoopPeriod)

	// waiting until all goroutines would finish
	time.Sleep(testControlLoopPeriod + delta)

	// no devices are onboarded in the system. Checking that after control loop the situation is the same
	ndList, err := grpcClient.GetDeviceList(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, ndList.GetDevices(), 0)

	// creating endpoints
	ep1 := server.CreateEndpoint(host1, port1, apiv1.Protocol_PROTOCOL_NETCONF)
	ep2 := server.CreateEndpoint(host2, port2, apiv1.Protocol_PROTOCOL_RESTCONF)
	ep3 := server.CreateEndpoint(host3, port3, apiv1.Protocol_PROTOCOL_SNMP)
	ep4 := server.CreateEndpoint(host4, port4, apiv1.Protocol_PROTOCOL_OPEN_V_SWITCH)
	// creating add network device request
	req1 := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_UBIQUITI, "XYZ", []*apiv1.Endpoint{ep1})
	req2 := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_UBIQUITI, "XYZ-nextgen", []*apiv1.Endpoint{ep2})
	req3 := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_CISCO, "xyz", []*apiv1.Endpoint{ep3})
	req4 := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_JUNIPER, "Zyx", []*apiv1.Endpoint{ep4})

	// adding all four network devices through API
	resp1, err := grpcClient.AddDevice(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, resp1)
	assert.True(t, resp1.GetAdded())
	resp2, err := grpcClient.AddDevice(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, resp2)
	assert.True(t, resp2.GetAdded())
	resp3, err := grpcClient.AddDevice(ctx, req3)
	require.NoError(t, err)
	require.NotNil(t, resp3)
	assert.True(t, resp3.GetAdded())
	resp4, err := grpcClient.AddDevice(ctx, req4)
	require.NoError(t, err)
	require.NotNil(t, resp4)
	assert.True(t, resp4.GetAdded())

	// now, four devices are onboarded in the system. Checking that this is what we have.
	ndList, err = grpcClient.GetDeviceList(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, ndList)
	assert.Len(t, ndList.GetDevices(), 4)

	// checking it also directly from the DB
	dbNDList, err := db.ListNetworkDevices(ctx, client)
	require.NoError(t, err)
	assert.NotEmpty(t, dbNDList)
	assert.Len(t, dbNDList, 4)

	// making sure that device statuses are empty (i.e., NOT present in the DB)
	dsReq1 := server.CreateGetDeviceStatusRequest(resp1.GetDevice().GetId(), ep1)
	dsReq2 := server.CreateGetDeviceStatusRequest(resp2.GetDevice().GetId(), ep2)
	dsReq3 := server.CreateGetDeviceStatusRequest(resp3.GetDevice().GetId(), ep3)
	dsReq4 := server.CreateGetDeviceStatusRequest(resp4.GetDevice().GetId(), ep4)

	// first device should return error - device status is not set
	retDS1, err := grpcClient.GetDeviceStatus(ctx, dsReq1)
	require.Error(t, err)
	require.Nil(t, retDS1)

	// second device should return error - device status is not set
	retDS2, err := grpcClient.GetDeviceStatus(ctx, dsReq2)
	require.Error(t, err)
	require.Nil(t, retDS2)

	// third device should return error - device status is not set
	retDS3, err := grpcClient.GetDeviceStatus(ctx, dsReq3)
	require.Error(t, err)
	require.Nil(t, retDS3)

	// fourth device should return error - device status is not set
	retDS4, err := grpcClient.GetDeviceStatus(ctx, dsReq4)
	require.Error(t, err)
	require.Nil(t, retDS4)

	// running another iteration of control loop
	sbManager.PerformControlLoopRoutine(testControlLoopPeriod)

	// waiting until all goroutines would finish
	time.Sleep(testControlLoopPeriod + delta)

	// checking that all devices are in UP state
	// first device should be with up status
	retDS1, err = grpcClient.GetDeviceStatus(ctx, dsReq1)
	require.NoError(t, err)
	require.NotNil(t, retDS1)
	assert.Equal(t, retDS1.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// second device should be with up status
	retDS2, err = grpcClient.GetDeviceStatus(ctx, dsReq2)
	require.NoError(t, err)
	require.NotNil(t, retDS2)
	assert.Equal(t, retDS2.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// third device should be with up status
	retDS3, err = grpcClient.GetDeviceStatus(ctx, dsReq3)
	require.NoError(t, err)
	require.NotNil(t, retDS3)
	assert.Equal(t, retDS3.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// fourth device should be with up status as well
	retDS4, err = grpcClient.GetDeviceStatus(ctx, dsReq4)
	require.NoError(t, err)
	require.NotNil(t, retDS4)
	assert.Equal(t, retDS4.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// Getting summary
	summary, err := grpcClient.GetSummary(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, summary.GetDevicesTotal(), int32(4))
	assert.Equal(t, summary.GetDevicesUp(), int32(4))
	assert.Equal(t, summary.GetDevicesUnhealthy(), int32(0))
	assert.Equal(t, summary.GetDownDevices(), int32(0))

	// now, changing environmental value, that is responsible for promoting Device State value to Network Device Simulator.
	t.Setenv(simulatorv1.EnvDeviceStatus, simulatorv1.DeviceStatusDOWN)
	// All device simulators, since they share the same environment, should report devices in down state
	// and return error on GetStatus operation. Counter tracking consequential connectivity absence should start increasing.
	// Right now, for the next 3 rounds, devices will remain in the same state, i.e., UP. Only on the 4th round
	// they will be reported in the DOWN state.

	// running another iteration of control loop
	sbManager.PerformControlLoopRoutine(testControlLoopPeriod)
	// waiting until all goroutines would finish
	time.Sleep(testControlLoopPeriod + delta)

	// repeating one more time
	sbManager.PerformControlLoopRoutine(testControlLoopPeriod)
	time.Sleep(testControlLoopPeriod + delta)

	// Checking that all devices are still in the UP state.
	// first device should be with up status
	retDS1, err = grpcClient.GetDeviceStatus(ctx, dsReq1)
	require.NoError(t, err)
	require.NotNil(t, retDS1)
	assert.Equal(t, retDS1.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// second device should be with up status
	retDS2, err = grpcClient.GetDeviceStatus(ctx, dsReq2)
	require.NoError(t, err)
	require.NotNil(t, retDS2)
	assert.Equal(t, retDS2.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// third device should be with up status
	retDS3, err = grpcClient.GetDeviceStatus(ctx, dsReq3)
	require.NoError(t, err)
	require.NotNil(t, retDS3)
	assert.Equal(t, retDS3.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// fourth device should be with up status as well
	retDS4, err = grpcClient.GetDeviceStatus(ctx, dsReq4)
	require.NoError(t, err)
	require.NotNil(t, retDS4)
	assert.Equal(t, retDS4.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_UP.String())

	// running another iteration of control loop
	sbManager.PerformControlLoopRoutine(testControlLoopPeriod)
	// waiting until all goroutines would finish
	time.Sleep(testControlLoopPeriod + delta)

	// Now, when threshold has passed, devices should be reported in DOWN state.
	// Checking that all devices are in DOWN state
	// First device should be with down status
	retDS1, err = grpcClient.GetDeviceStatus(ctx, dsReq1)
	require.NoError(t, err)
	require.NotNil(t, retDS1)
	assert.Equal(t, retDS1.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_DOWN.String())

	// second device should be with down status
	retDS2, err = grpcClient.GetDeviceStatus(ctx, dsReq2)
	require.NoError(t, err)
	require.NotNil(t, retDS2)
	assert.Equal(t, retDS2.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_DOWN.String())

	// third device should be with down status
	retDS3, err = grpcClient.GetDeviceStatus(ctx, dsReq3)
	require.NoError(t, err)
	require.NotNil(t, retDS3)
	assert.Equal(t, retDS3.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_DOWN.String())

	// fourth device should be with down status as well
	retDS4, err = grpcClient.GetDeviceStatus(ctx, dsReq4)
	require.NoError(t, err)
	require.NotNil(t, retDS4)
	assert.Equal(t, retDS4.GetStatus().GetStatus().String(), apiv1.Status_STATUS_DEVICE_DOWN.String())

	// Getting fresh summary
	summary, err = grpcClient.GetSummary(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, summary.GetDevicesTotal(), int32(4))
	assert.Equal(t, summary.GetDevicesUp(), int32(0))
	assert.Equal(t, summary.GetDevicesUnhealthy(), int32(0))
	assert.Equal(t, summary.GetDownDevices(), int32(4))

	// removing all devices from the system
	delResp1, err := grpcClient.DeleteDevice(ctx, server.CreateDeleteDeviceRequest(resp1.GetDevice().GetId()))
	require.NoError(t, err)
	require.NotNil(t, delResp1)
	assert.True(t, delResp1.GetDeleted())

	delResp2, err := grpcClient.DeleteDevice(ctx, server.CreateDeleteDeviceRequest(resp2.GetDevice().GetId()))
	require.NoError(t, err)
	require.NotNil(t, delResp2)
	assert.True(t, delResp2.GetDeleted())

	delResp3, err := grpcClient.DeleteDevice(ctx, server.CreateDeleteDeviceRequest(resp3.GetDevice().GetId()))
	require.NoError(t, err)
	require.NotNil(t, delResp3)
	assert.True(t, delResp3.GetDeleted())

	delResp4, err := grpcClient.DeleteDevice(ctx, server.CreateDeleteDeviceRequest(resp4.GetDevice().GetId()))
	require.NoError(t, err)
	require.NotNil(t, delResp4)
	assert.True(t, delResp4.GetDeleted())
}
