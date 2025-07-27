// Package server_test implements unit tests for gRPC server functions.
package server_test

import (
	"context"
	"os"
	"testing"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var grpcClient apiv1.DeviceMonitoringServiceClient

func TestMain(m *testing.M) {
	var err error
	entClient, serverClient, wg, termChan, reverseProxyTermChan, err := monitoring_testing.SetupFull()
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
