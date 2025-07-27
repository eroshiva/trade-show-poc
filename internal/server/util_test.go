// Package server_test provides unit tests for server package.
package server_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	deviceModel     = "XYZ"
	deviceHwVersion = "HW-XYZ"
	deviceVendor    = networkdevice.VendorVENDOR_UBIQUITI
	host1           = "192.168.0.1"
	port1           = "123"
	protocol1       = endpoint.ProtocolPROTOCOL_NETCONF

	host2     = "192.168.0.2"
	port2     = "456"
	protocol2 = endpoint.ProtocolPROTOCOL_SNMP

	host3     = "192.168.0.3"
	port3     = "789"
	protocol3 = endpoint.ProtocolPROTOCOL_RESTCONF
)

var (
	client     *ent.Client
	version    = "XYZ"
	checksum   = fmt.Sprintf("%x", sha256.Sum256([]byte(version)))
	fwVersion  = "fw-" + version
	fwChecksum = fmt.Sprintf("%x", sha256.Sum256([]byte(fwVersion)))
)

func TestConvertNetworkDeviceResourceToNetworkDeviceProto(t *testing.T) {
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

	ep2, err := db.CreateEndpoint(ctx, client, host2, port2, protocol2)
	require.NoError(t, err)
	require.NotNil(t, ep2)
	t.Cleanup(func() {
		err = db.DeleteEndpointByID(ctx, client, ep2.ID)
		assert.NoError(t, err)
	})

	// creating SW and FW versions
	sw, err := db.CreateVersion(ctx, client, version, checksum)
	require.NoError(t, err)
	require.NotNil(t, sw)
	t.Cleanup(func() {
		err = db.DeleteVersionByID(ctx, client, sw.ID)
		assert.NoError(t, err)
	})

	fw, err := db.CreateVersion(ctx, client, fwVersion, fwChecksum)
	require.NoError(t, err)
	require.NotNil(t, fw)
	t.Cleanup(func() {
		err = db.DeleteVersionByID(ctx, client, fw.ID)
		assert.NoError(t, err)
	})

	// creating network device
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep1, ep2})
	require.NoError(t, err)
	require.NotNil(t, nd)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
		assert.NoError(t, err)
	})

	// adding HW, SW, and FW versions to the network device
	updND, err := db.UpdateNetworkDeviceVersions(ctx, client, nd.ID, deviceHwVersion, sw, fw)
	require.NoError(t, err)
	require.NotNil(t, updND)

	// converting network device to proto notation of network device
	protoND := server.ConvertNetworkDeviceResourceToNetworkDeviceProto(updND)
	assert.Equal(t, updND.ID, protoND.GetId())
	assert.Equal(t, updND.Model, protoND.GetModel())
	assert.Equal(t, server.ConvertEntVendorToProtoVendor(updND.Vendor), protoND.GetVendor())
	assert.Equal(t, updND.HwVersion, protoND.GetHwVersion())

	require.NotNil(t, updND.Edges.Endpoints)
	require.NotNil(t, protoND.GetEndpoints())
	assert.Len(t, updND.Edges.Endpoints, len(protoND.GetEndpoints()))

	// asserting endpoints
	for _, ep := range updND.Edges.Endpoints {
		protoEP := server.ConvertEndpointToEndpointProto(ep)
		require.NotNil(t, protoEP)

		assert.Equal(t, ep.ID, protoEP.GetId())
		assert.Equal(t, ep.Host, protoEP.GetHost())
		assert.Equal(t, ep.Port, protoEP.GetPort())
		assert.Equal(t, server.ConvertEntProtocolToProtoProtocol(ep.Protocol), protoEP.GetProtocol())
	}

	require.NotNil(t, updND.Edges.SwVersion)
	require.NotNil(t, protoND.GetSwVersion())
	assert.Equal(t, updND.Edges.SwVersion.Version, protoND.GetSwVersion().GetVersion())
	assert.Equal(t, updND.Edges.SwVersion.Checksum, protoND.GetSwVersion().GetChecksum())

	require.NotNil(t, updND.Edges.FwVersion)
	require.NotNil(t, protoND.GetFwVersion())
	assert.Equal(t, updND.Edges.FwVersion.Version, protoND.GetFwVersion().GetVersion())
	assert.Equal(t, updND.Edges.FwVersion.Checksum, protoND.GetFwVersion().GetChecksum())
}
