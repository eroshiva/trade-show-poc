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
	deviceModel  = "XYZ"
	deviceVendor = networkdevice.VendorVENDOR_UBIQUITI
	host1        = "192.168.0.1"
	port1        = "123"
	protocol1    = endpoint.ProtocolPROTOCOL_NETCONF

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
	err = monitoring_testing.GracefullyCloseEntClient(client)
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

	// retrieving all version resources and making sure there is only two of them
	vs, err := db.ListVersions(ctx, client)
	assert.NoError(t, err)
	assert.Len(t, vs, 2)

	// retrieving back resource and comparing it to the original one
	retV, err := db.GetVersionByID(ctx, client, v.ID)
	assert.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v, retV)

	// updating version
	v.Version = version + "-new"
	v.Checksum = fmt.Sprintf("%x", sha256.Sum256([]byte(v.Version)))
	updV, err := db.UpdateVersion(ctx, client, v.ID, v.Version, v.Checksum)
	require.NoError(t, err)
	monitoring_testing.AssertEqualVersion(t, v, updV)

	// Deleting version resource
	err = db.DeleteVersionByID(ctx, client, v.ID)
	assert.NoError(t, err)

	// retrieving all version resources and making sure there is only one
	vs, err = db.ListVersions(ctx, client)
	assert.NoError(t, err)
	assert.Len(t, vs, 1)
}

func TestVersionResourceErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoring_testing.DefaultTestTimeout)
	t.Cleanup(cancel)

	// creating Version resource
	v, err := db.CreateVersion(ctx, client, version, checksum)
	require.NoError(t, err)
	require.NotNil(t, v)
	// Cleaning up version resource at the end of the test
	t.Cleanup(func() {
		err = db.DeleteVersionByID(ctx, client, v.ID)
		assert.NoError(t, err)
	})

	// failing get
	_, err = db.GetVersionByID(ctx, client, uuid.New().String())
	assert.Error(t, err)

	// failing update
	_, err = db.UpdateVersion(ctx, client, uuid.New().String(), v.Version, v.Checksum)
	require.Error(t, err)

	// failing delete
	_ = db.DeleteVersionByID(ctx, client, uuid.New().String())

	// retrieving all version resources and making sure there is only one (initial one)
	vs, err := db.ListVersions(ctx, client)
	assert.NoError(t, err)
	assert.Len(t, vs, 1)
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
	ds, err := db.CreateDeviceStatus(ctx, client, devicestatus.StatusSTATUS_DEVICE_UP, time.Now().String(), nd)
	require.NoError(t, err)
	require.NotNil(t, ds)

	// retrieving device status back
	retDs, err := db.GetDeviceStatusByID(ctx, client, ds.ID)
	require.NoError(t, err)
	monitoring_testing.AssertDeviceStatus(t, ds, retDs)

	// updating network device status
	ds.Status = devicestatus.StatusSTATUS_DEVICE_UNHEALTHY
	ds.LastSeen = time.Now().String()
	updDs, err := db.UpdateDeviceStatusByNetworkDeviceID(ctx, client, nd.ID, ds.Status, ds.LastSeen)
	require.NoError(t, err)
	require.NotNil(t, updDs)
	monitoring_testing.AssertDeviceStatus(t, ds, updDs)

	time.Sleep(200 * time.Millisecond)

	// retrieving device status by Network Device ID
	retDs2, err := db.GetDeviceStatusByNetworkDeviceID(ctx, client, nd.ID)
	require.NoError(t, err)
	require.NotNil(t, retDs2)
	monitoring_testing.AssertDeviceStatus(t, ds, retDs2)

	// updating network device status again
	ds.Status = devicestatus.StatusSTATUS_DEVICE_DOWN
	updDs2, err := db.UpdateDeviceStatusByEndpointID(ctx, client, ep2.ID, ds.Status, "")
	require.NoError(t, err)
	require.NotNil(t, updDs2)
	monitoring_testing.AssertDeviceStatus(t, ds, updDs2)

	// retrieving device status by Endpoint ID
	retDs3, err := db.GetDeviceStatusByEndpointID(ctx, client, ep2.ID)
	require.NoError(t, err)
	require.NotNil(t, retDs3)
	monitoring_testing.AssertDeviceStatus(t, ds, retDs3)

	// listing all device statuses present in the system, there should be only one
	listDs, err := db.ListDeviceStatusResources(ctx, client)
	require.NoError(t, err)
	require.NotNil(t, listDs)
	assert.Len(t, listDs, 1)

	// removing device status from the DB
	err = db.DeleteDeviceStatusByID(ctx, client, ds.ID)
	assert.NoError(t, err)
}

func TestNetworkDeviceResource(t *testing.T) {
	// creating Network Device resource
}
