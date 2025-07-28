// Package db_test provides unit tests for DB client interation with DB.
package db_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	monitoring_testing "github.com/eroshiva/trade-show-poc/pkg/testing"
	"github.com/google/uuid"
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
)

var (
	client   *ent.Client
	version  = "XYZ"
	checksum = fmt.Sprintf("%x", sha256.Sum256([]byte(version)))
)

func TestMain(m *testing.M) {
	var err error
	client, err = monitoring_testing.Setup()
	if err != nil {
		panic(err)
	}
	code := m.Run()
	err = db.GracefullyCloseDBClient(client)
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestVersionResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating Version resource
	v, err := db.CreateVersion(ctx, client, version, checksum)
	require.NoError(t, err)
	require.NotNil(t, v)

	// creating another version resource
	v2, err := db.CreateVersion(ctx, client, version, checksum)
	require.NoError(t, err)
	require.NotNil(t, v2)
	t.Cleanup(func() {
		err = db.DeleteVersionByID(ctx, client, v2.ID)
		assert.NoError(t, err)
	})

	// retrieving back resource and comparing it to the original one
	retV, err := db.GetVersionByID(ctx, client, v.ID)
	assert.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v, retV)

	// retrieving back resource and comparing it to the original one
	retV2, err := db.GetVersionByID(ctx, client, v2.ID)
	assert.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v2, retV2)

	// updating version
	v.Version = version + "-new"
	v.Checksum = fmt.Sprintf("%x", sha256.Sum256([]byte(v.Version)))
	updV, err := db.UpdateVersion(ctx, client, v.ID, v.Version, v.Checksum)
	require.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v, updV)

	// Deleting version resource
	err = db.DeleteVersionByID(ctx, client, v.ID)
	assert.NoError(t, err)

	// retrieving back first version resource - should fail, version resource is gone
	retV, err = db.GetVersionByID(ctx, client, v.ID)
	require.Error(t, err)
	require.Nil(t, retV)

	// retrieving back second version resource and comparing it to the original one
	retV2, err = db.GetVersionByID(ctx, client, v2.ID)
	assert.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v2, retV2)
}

func TestVersionResourceErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// fail creating Version resource with invalid data
	v, err := db.CreateVersion(ctx, client, "", checksum)
	require.Error(t, err)
	require.Nil(t, v)

	// creating Version resource
	v, err = db.CreateVersion(ctx, client, version, checksum)
	require.NoError(t, err)
	require.NotNil(t, v)
	// Cleaning up version resource at the end of the test
	t.Cleanup(func() {
		err = db.DeleteVersionByID(ctx, client, v.ID)
		assert.NoError(t, err)
	})

	// failing get
	_, err = db.GetVersionByID(ctx, client, uuid.NewString())
	assert.Error(t, err)

	// failing update
	_, err = db.UpdateVersion(ctx, client, uuid.NewString(), v.Version, v.Checksum)
	require.Error(t, err)

	// failing delete
	_ = db.DeleteVersionByID(ctx, client, uuid.NewString())

	// retrieving initial version resource and making sure this is exactly the one that's been created at the beginning
	retV, err := db.GetVersionByID(ctx, client, v.ID)
	assert.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v, retV)
}

func TestDeviceStatusResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating two endpoints
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

	// creating network device resource
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep1, ep2})
	require.NoError(t, err)
	require.NotNil(t, nd)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
		assert.NoError(t, err)
	})

	// creating device status
	ds, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_UP, time.Now().String(), 0, nd)
	require.NoError(t, err)
	require.NotNil(t, ds)

	// retrieving device status back
	retDs, err := db.GetDeviceStatusByID(ctx, client, ds.ID)
	require.NoError(t, err)
	monitoring_testing.AssertDeviceStatus(t, ds, retDs)

	// updating network device status
	ds.Status = devicestatus.StatusSTATUS_DEVICE_UNHEALTHY
	ds.LastSeen = time.Now().String()
	updDs, err := db.UpdateDeviceStatusByNetworkDeviceID(ctx, client, nd.ID, ds.Status, ds.LastSeen, 0)
	require.NoError(t, err)
	require.NotNil(t, updDs)
	monitoring_testing.AssertDeviceStatus(t, ds, updDs)

	// retrieving device status by Network Device ID
	retDs2, err := db.GetDeviceStatusByNetworkDeviceID(ctx, client, nd.ID)
	require.NoError(t, err)
	require.NotNil(t, retDs2)
	monitoring_testing.AssertDeviceStatus(t, ds, retDs2)

	// updating network device status again
	ds.Status = devicestatus.StatusSTATUS_DEVICE_DOWN
	updDs2, err := db.UpdateDeviceStatusByEndpointID(ctx, client, ep2.ID, ds.Status, "", 0)
	require.NoError(t, err)
	require.NotNil(t, updDs2)
	monitoring_testing.AssertDeviceStatus(t, ds, updDs2)

	// retrieving device status by Endpoint ID
	retDs3, err := db.GetDeviceStatusByEndpointID(ctx, client, ep2.ID)
	require.NoError(t, err)
	require.NotNil(t, retDs3)
	monitoring_testing.AssertDeviceStatus(t, ds, retDs3)

	// removing device status from the DB
	err = db.DeleteDeviceStatusByID(ctx, client, ds.ID)
	assert.NoError(t, err)
}

func TestDeviceStatusResourceErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// fail - creating device status with no network device
	ds, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_UP, time.Now().String(), 0, nil)
	require.Error(t, err)
	require.Nil(t, ds)

	// retrieving non-existing device status
	retDs, err := db.GetDeviceStatusByID(ctx, client, uuid.NewString())
	require.Error(t, err)
	assert.Nil(t, retDs)

	// retrieving non-existing device status by network device ID
	retDs2, err := db.GetDeviceStatusByNetworkDeviceID(ctx, client, uuid.NewString())
	require.Error(t, err)
	assert.Nil(t, retDs2)

	// retrieving non-existing device status by endpoint ID
	retDs3, err := db.GetDeviceStatusByEndpointID(ctx, client, uuid.NewString())
	require.Error(t, err)
	assert.Nil(t, retDs3)

	// creating two endpoints
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

	// creating network device resource
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{ep1, ep2})
	require.NoError(t, err)
	require.NotNil(t, nd)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
		assert.NoError(t, err)
	})

	// fail - creating device status with no status
	ds2, err := db.CreateDeviceStatus(ctx, client, "", time.Now().String(), 0, nd)
	require.Error(t, err)
	require.Nil(t, ds2)

	// fail - updating network device status by fake network device ID
	status := devicestatus.StatusSTATUS_DEVICE_UP
	lastSeen := time.Now().String()
	updDs, err := db.UpdateDeviceStatusByNetworkDeviceID(ctx, client, uuid.NewString(), status, lastSeen, 0)
	assert.Error(t, err)
	assert.Nil(t, updDs)

	// success - creating network device status on update
	updDs, err = db.UpdateDeviceStatusByNetworkDeviceID(ctx, client, nd.ID, status, lastSeen, 0)
	require.NoError(t, err)
	require.NotNil(t, updDs)

	// deleting device status
	err = db.DeleteDeviceStatusByID(ctx, client, updDs.ID)
	assert.NoError(t, err)

	// fail - updating network device status by fake endpoint ID
	status = devicestatus.StatusSTATUS_DEVICE_UNHEALTHY
	lastSeen = time.Now().String()
	updDs, err = db.UpdateDeviceStatusByEndpointID(ctx, client, uuid.NewString(), status, lastSeen, 0)
	assert.Error(t, err)
	assert.Nil(t, updDs)

	// success - creating network device status on update
	updDs, err = db.UpdateDeviceStatusByEndpointID(ctx, client, ep2.ID, status, lastSeen, 0)
	require.NoError(t, err)
	require.NotNil(t, updDs)

	// deleting device status
	err = db.DeleteDeviceStatusByID(ctx, client, updDs.ID)
	assert.NoError(t, err)
}

func TestEndpointResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating endpoint
	ep, err := db.CreateEndpoint(ctx, client, host1, port1, protocol1)
	require.NoError(t, err)
	require.NotNil(t, ep)

	// retrieving endpoint back
	retEp, err := db.GetEndpointByID(ctx, client, ep.ID)
	require.NoError(t, err)
	require.NotNil(t, retEp)
	monitoring_testing.AssertEqualEndpoints(t, ep, retEp)

	// updating endpoint
	ep.Port = port2
	ep.Protocol = protocol2
	updEp, err := db.UpdateEndpoint(ctx, client, ep.ID, ep.Host, ep.Port, ep.Protocol)
	require.NoError(t, err)
	require.NotNil(t, updEp)
	monitoring_testing.AssertEqualEndpoints(t, ep, updEp)

	// deleting endpoint
	err = db.DeleteEndpointByID(ctx, client, updEp.ID)
	assert.NoError(t, err)
}

func TestEndpointResourceErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// fail - creating endpoint with incomplete parameters
	ep, err := db.CreateEndpoint(ctx, client, host1, port1, "")
	require.Error(t, err)
	require.Nil(t, ep)

	// fail - retrieving endpoint with fake ID
	retEp, err := db.GetEndpointByID(ctx, client, uuid.NewString())
	require.Error(t, err)
	require.Nil(t, retEp)

	// fail - updating endpoint with fake UUID
	updEp, err := db.UpdateEndpoint(ctx, client, uuid.NewString(), "", "", "")
	require.Error(t, err)
	require.Nil(t, updEp)
}

func TestNetworkDeviceResource(t *testing.T) {
	// creating Network Device resource
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating network device resource with no endpoints
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd)

	// retrieving back network device resource
	retNd, err := db.GetNetworkDeviceByID(ctx, client, nd.ID)
	require.NoError(t, err)
	require.NotNil(t, retNd)
	monitoring_testing.AssertEqualNetworkDevicesNoEdges(t, nd, retNd)

	// listing all network devices present in the system
	listNd, err := db.ListNetworkDevices(ctx, client)
	require.NoError(t, err)
	require.NotNil(t, listNd)
	assert.Len(t, listNd, 1)

	// creating version resources
	sw, err := db.CreateVersion(ctx, client, version, checksum)
	require.NoError(t, err)
	require.NotNil(t, sw)

	fwVersion := "fw-" + version
	fwChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(fwVersion)))
	fw, err := db.CreateVersion(ctx, client, fwVersion, fwChecksum)
	require.NoError(t, err)
	require.NotNil(t, fw)

	// updating network device resource with new versions
	nd.HwVersion = deviceHwVersion
	nd.Edges.Endpoints = make([]*ent.Endpoint, 0)
	nd.Edges.SwVersion = sw
	nd.Edges.FwVersion = fw
	updNd, err := db.UpdateNetworkDeviceVersions(ctx, client, nd.ID, deviceHwVersion, sw, fw)
	require.NoError(t, err)
	require.NotNil(t, updNd)
	monitoring_testing.AssertEqualNetworkDevices(t, nd, updNd)

	err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
	assert.NoError(t, err)
}

func TestNetworkDeviceResourceEndpoints(t *testing.T) {
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

	// creating network device
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd)

	// adding endpoint to network device
	nd.Edges.Endpoints = []*ent.Endpoint{ep1}
	updNd, err := db.UpdateNetworkDeviceAddEndpoints(ctx, client, nd.ID, []*ent.Endpoint{ep1})
	require.NoError(t, err)
	require.NotNil(t, updNd)
	monitoring_testing.AssertEqualNetworkDevicesEndpointsOnly(t, nd, updNd)

	// substituting endpoints
	nd.Edges.Endpoints = []*ent.Endpoint{ep2}
	updNd2, err := db.UpdateNetworkDeviceEndpoints(ctx, client, nd.ID, []*ent.Endpoint{ep2})
	require.NoError(t, err)
	require.NotNil(t, updNd2)
	monitoring_testing.AssertEqualNetworkDevicesEndpointsOnly(t, nd, updNd2)

	// overwriting endpoints again
	nd.Edges.Endpoints = []*ent.Endpoint{ep1}
	updNd3, err := db.UpdateNetworkDeviceByUser(ctx, client, nd.ID, deviceModel, deviceVendor, []*ent.Endpoint{ep1})
	require.NoError(t, err)
	require.NotNil(t, updNd)
	monitoring_testing.AssertEqualNetworkDevicesEndpointsOnly(t, nd, updNd3)

	// deleting network device
	err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
	assert.NoError(t, err)
}

func TestNetworkDeviceResourceErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// fail - incomplete network device data
	nd, err := db.CreateNetworkDevice(ctx, client, deviceModel, "", []*ent.Endpoint{})
	require.Error(t, err)
	require.Nil(t, nd)

	// creating a network device
	nd, err = db.CreateNetworkDevice(ctx, client, deviceModel, deviceVendor, []*ent.Endpoint{})
	require.NoError(t, err)
	require.NotNil(t, nd)
	t.Cleanup(func() {
		err = db.DeleteNetworkDeviceByID(ctx, client, nd.ID)
		assert.NoError(t, err)
	})

	// fail - updating network device with fake UUID
	updNd, err := db.UpdateNetworkDeviceByUser(ctx, client, uuid.NewString(), deviceModel, deviceVendor, []*ent.Endpoint{})
	require.Error(t, err)
	require.Nil(t, updNd)

	// fail - updating network device with fake UUID
	updNd2, err := db.UpdateNetworkDeviceEndpoints(ctx, client, uuid.NewString(), []*ent.Endpoint{})
	require.Error(t, err)
	require.Nil(t, updNd2)

	// fail - updating network device with an empty list of endpoints
	updNd3, err := db.UpdateNetworkDeviceEndpoints(ctx, client, nd.ID, []*ent.Endpoint{})
	require.Error(t, err)
	require.Nil(t, updNd3)

	// fail - updating network device with fake UUID
	updNd4, err := db.UpdateNetworkDeviceAddEndpoints(ctx, client, uuid.NewString(), []*ent.Endpoint{})
	require.Error(t, err)
	require.Nil(t, updNd4)

	// fail - updating network device with an empty list of endpoints
	updNd5, err := db.UpdateNetworkDeviceAddEndpoints(ctx, client, uuid.NewString(), []*ent.Endpoint{})
	require.Error(t, err)
	require.Nil(t, updNd5)

	// fail - get network device by not existing endpoint
	retNd, err := db.GetNetworkDeviceByEndpoint(ctx, client, host1, port2)
	require.Error(t, err)
	require.Nil(t, retNd)
}
