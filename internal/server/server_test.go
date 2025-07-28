// Package server_test implements unit tests for gRPC server functions.
package server_test

import (
	"context"
	"os"
	"testing"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var grpcClient apiv1.DeviceMonitoringServiceClient

func TestMain(m *testing.M) {
	var err error
	entClient, serverClient, wg, termChan, reverseProxyTermChan, err := monitoring_testing.SetupFull("", "")
	if err != nil {
		panic(err)
	}
	client = entClient
	grpcClient = serverClient

	// running tests
	code := m.Run()

	// all tests were run, stopping servers gracefully
	monitoring_testing.TeardownFull(client, wg, termChan, reverseProxyTermChan)
	os.Exit(code)
}

func TestAddDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating endpoints
	endpoints := make([]*apiv1.Endpoint, 0)
	ep1 := server.CreateEndpoint(host1, port1, apiv1.Protocol_PROTOCOL_NETCONF)
	ep2 := server.CreateEndpoint(host2, port2, apiv1.Protocol_PROTOCOL_NETCONF)
	endpoints = append(endpoints, ep1, ep2)
	// creating request
	req := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_UBIQUITI, "XYZ", endpoints)

	// sending request to the server
	res, err := grpcClient.AddDevice(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.GetAdded())
	assert.Empty(t, res.GetDetails())

	// checking that DB has created new Network Device resource
	nd, err := db.GetNetworkDeviceByID(ctx, client, res.GetDevice().GetId())
	require.NoError(t, err)
	require.NotNil(t, nd)
	t.Cleanup(func() {
		// removing network device
		err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
		assert.NoError(t, err)
	})

	// running limited set of assertions
	assert.Equal(t, res.GetDevice().GetId(), nd.ID)
	assert.Equal(t, res.GetDevice().GetVendor(), server.ConvertEntVendorToProtoVendor(nd.Vendor))
	assert.Equal(t, res.GetDevice().GetModel(), nd.Model)
	assert.Len(t, res.GetDevice().GetEndpoints(), 2)
	assert.Len(t, nd.Edges.Endpoints, 2)
}

func TestDeleteDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating network device resource with no endpoints
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd)

	// removing network device via API
	resp, err := grpcClient.DeleteDevice(ctx, server.CreateDeleteDeviceRequest(nd.ID))
	assert.NoError(t, err)
	assert.Equal(t, resp.GetId(), nd.ID)
	assert.True(t, resp.GetDeleted())
	assert.Empty(t, resp.GetDetails())
}

func TestGetDeviceList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating network device resource with no endpoints
	nd1, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd1)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd1.ID)
		assert.NoError(t, err)
	})

	nd2, err := db.CreateNetworkDevice(ctx, client, deviceModel+"-new", deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd2)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd2.ID)
		assert.NoError(t, err)
	})

	ndList, err := grpcClient.GetDeviceList(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, ndList)
	assert.Len(t, ndList.GetDevices(), 2)
}

func TestUpdateNetworkDevice(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating endpoints
	ep1, err := db.CreateEndpoint(ctx, client, host1, port1, protocol1)
	require.NoError(t, err)
	require.NotNil(t, ep1)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep1.ID)
		assert.NoError(t, err)
	})
	// converting endpoint resource to Proto notation
	ep1Proto := server.ConvertEndpointToEndpointProto(ep1)

	ep2, err := db.CreateEndpoint(ctx, client, host2, port2, protocol2)
	require.NoError(t, err)
	require.NotNil(t, ep2)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep2.ID)
		assert.NoError(t, err)
	})
	// converting endpoint resource to proto notation
	ep2Proto := server.ConvertEndpointToEndpointProto(ep2)

	// creating network device resource with no endpoints
	nd1, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd1)
	assert.Nil(t, nd1.Edges.Endpoints) // make sure there is no endpoint
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd1.ID)
		assert.NoError(t, err)
	})

	nd2, err := db.CreateNetworkDevice(ctx, client, deviceModel+"-new", deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd2)
	assert.Nil(t, nd2.Edges.Endpoints) // make sure there is no endpoint
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd2.ID)
		assert.NoError(t, err)
	})

	// converting network device resources to Proto notation
	nd1Proto := server.ConvertNetworkDeviceResourceToNetworkDeviceProto(nd1)
	nd2Proto := server.ConvertNetworkDeviceResourceToNetworkDeviceProto(nd2)

	// adding endpoints to the network devices
	nd1Proto.Endpoints = append(nd1Proto.Endpoints, ep1Proto)
	nd2Proto.Endpoints = append(nd2Proto.Endpoints, ep2Proto)

	// updating network device to have endpoints from NB API
	retList, err := grpcClient.UpdateDeviceList(ctx, server.CreateUpdateDeviceListRequest([]*apiv1.NetworkDevice{nd1Proto, nd2Proto}))
	require.NoError(t, err)
	require.NotNil(t, retList)
	assert.Len(t, retList.GetDevices(), 2)

	// check that Network Devices were updated with endpoints
	nd1, err = db.GetNetworkDeviceByID(ctx, client, nd1.ID)
	require.NoError(t, err)
	require.NotNil(t, nd1)
	assert.NotNil(t, nd1.Edges.Endpoints)

	nd2, err = db.GetNetworkDeviceByID(ctx, client, nd2.ID)
	require.NoError(t, err)
	require.NotNil(t, nd2)
	assert.NotNil(t, nd2.Edges.Endpoints)
}

func TestGetDeviceStatus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating endpoint
	ep1, err := db.CreateEndpoint(ctx, client, host1, port1, protocol1)
	require.NoError(t, err)
	require.NotNil(t, ep1)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep1.ID)
		assert.NoError(t, err)
	})
	// converting endpoint resource to Proto notation
	ep1Proto := server.ConvertEndpointToEndpointProto(ep1)

	// creating network device resource with no endpoints
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep1})
	require.NoError(t, err)
	require.NotNil(t, nd)
	assert.NotNil(t, nd.Edges.Endpoints) // make sure there is one endpoint
	assert.Len(t, nd.Edges.Endpoints, 1)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
		assert.NoError(t, err)
	})

	// creating device status
	timestamp := time.Now().String()
	ds, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_UP, timestamp, 0, nd)
	require.NoError(t, err)
	require.NotNil(t, ds)
	t.Cleanup(func() {
		err = db.DeleteDeviceStatusByID(ctx, client, ds.ID)
		assert.NoError(t, err)
	})

	// retrieving device status from the NB API
	req := server.CreateGetDeviceStatusRequest(nd.ID, ep1Proto)
	retDS, err := grpcClient.GetDeviceStatus(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, retDS)

	// running few assertions
	assert.Equal(t, retDS.GetStatus().GetId(), ds.ID)
	assert.Equal(t, retDS.GetStatus().GetStatus(), server.ConvertEntStatusToProtoStatus(ds.Status))
	assert.Equal(t, retDS.GetStatus().GetLastSeen(), timestamp)
}

func TestGetSummary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating endpoint
	ep1, err := db.CreateEndpoint(ctx, client, host1, port1, protocol1)
	require.NoError(t, err)
	require.NotNil(t, ep1)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep1.ID)
		assert.NoError(t, err)
	})

	// creating network device resource with no endpoints
	nd1, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep1})
	require.NoError(t, err)
	require.NotNil(t, nd1)
	assert.NotNil(t, nd1.Edges.Endpoints) // make sure there is one endpoint
	assert.Len(t, nd1.Edges.Endpoints, 1)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd1.ID)
		assert.NoError(t, err)
	})

	// creating device status
	timestamp1 := time.Now().String()
	ds1, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_UP, timestamp1, 0, nd1)
	require.NoError(t, err)
	require.NotNil(t, ds1)
	t.Cleanup(func() {
		err = db.DeleteDeviceStatusByID(ctx, client, ds1.ID)
		assert.NoError(t, err)
	})

	// creating endpoint
	ep2, err := db.CreateEndpoint(ctx, client, host2, port2, protocol2)
	require.NoError(t, err)
	require.NotNil(t, ep2)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep2.ID)
		assert.NoError(t, err)
	})

	// creating network device resource with no endpoints
	nd2, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep2})
	require.NoError(t, err)
	require.NotNil(t, nd2)
	assert.NotNil(t, nd2.Edges.Endpoints) // make sure there is one endpoint
	assert.Len(t, nd2.Edges.Endpoints, 1)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd2.ID)
		assert.NoError(t, err)
	})

	// creating device status
	timestamp2 := time.Now().String()
	ds2, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_DOWN, timestamp2, 0, nd2)
	require.NoError(t, err)
	require.NotNil(t, ds2)
	t.Cleanup(func() {
		err = db.DeleteDeviceStatusByID(ctx, client, ds2.ID)
		assert.NoError(t, err)
	})

	// creating endpoint
	ep3, err := db.CreateEndpoint(ctx, client, host3, port3, protocol3)
	require.NoError(t, err)
	require.NotNil(t, ep3)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep3.ID)
		assert.NoError(t, err)
	})

	// creating network device resource with no endpoints
	nd3, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep3})
	require.NoError(t, err)
	require.NotNil(t, nd3)
	assert.NotNil(t, nd3.Edges.Endpoints) // make sure there is one endpoint
	assert.Len(t, nd3.Edges.Endpoints, 1)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd3.ID)
		assert.NoError(t, err)
	})

	// creating device status
	timestamp3 := time.Now().String()
	ds3, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_UNHEALTHY, timestamp3, 0, nd3)
	require.NoError(t, err)
	require.NotNil(t, ds3)
	t.Cleanup(func() {
		err = db.DeleteDeviceStatusByID(ctx, client, ds3.ID)
		assert.NoError(t, err)
	})

	// retrieving summary
	summary, err := grpcClient.GetSummary(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, summary)

	// running assertions
	assert.Equal(t, summary.GetDevicesTotal(), int32(3))
	assert.Equal(t, summary.GetDevicesUp(), int32(1))
	assert.Equal(t, summary.GetDownDevices(), int32(1))
	assert.Equal(t, summary.GetDevicesUnhealthy(), int32(1))
}

func TestSwapDeviceList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating endpoints
	ep1 := server.CreateEndpoint(host1, port1, apiv1.Protocol_PROTOCOL_NETCONF)
	ep2 := server.CreateEndpoint(host2, port2, apiv1.Protocol_PROTOCOL_NETCONF)
	// creating AddDevice requests
	req1 := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_UBIQUITI, "XYZ", []*apiv1.Endpoint{ep1})
	req2 := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_UBIQUITI, "XYZ", []*apiv1.Endpoint{ep2})

	// sending request to the server
	res1, err := grpcClient.AddDevice(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, res1)
	assert.True(t, res1.GetAdded())
	assert.Empty(t, res1.GetDetails())

	// checking that DB has created new Network Device resource
	nd1, err := db.GetNetworkDeviceByEndpoint(ctx, client, ep1.GetHost(), ep1.GetPort())
	require.NoError(t, err)
	require.NotNil(t, nd1)

	// sending another request to the server
	res2, err := grpcClient.AddDevice(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, res2)
	assert.True(t, res2.GetAdded())
	assert.Empty(t, res2.GetDetails())

	// checking that DB has created new Network Device resource
	nd2, err := db.GetNetworkDeviceByEndpoint(ctx, client, ep2.GetHost(), ep2.GetPort())
	require.NoError(t, err)
	require.NotNil(t, nd2)

	// performing swap with only one device in the list
	protoND1 := server.ConvertNetworkDeviceResourceToNetworkDeviceProto(nd1)
	resp, err := grpcClient.SwapDeviceList(ctx, &apiv1.SwapDeviceListRequest{
		Devices: []*apiv1.NetworkDevice{protoND1},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.GetDevices(), 1)

	// checking that DB still has first network device resource
	nd1, err = db.GetNetworkDeviceByEndpoint(ctx, client, ep1.GetHost(), ep1.GetPort())
	require.NoError(t, err)
	require.NotNil(t, nd1)
	t.Cleanup(func() {
		// removing network device
		err = db.DeleteNetworkDeviceByID(ctx, client, nd1.ID)
		assert.NoError(t, err)
	})

	// checking that second network device does not exist in the DB anymore
	nd2, err = db.GetNetworkDeviceByEndpoint(ctx, client, ep2.GetHost(), ep2.GetPort())
	require.Error(t, err)
	require.Nil(t, nd2)
}
