// Package testing contains helper functions to run unit tests within this repository.
package testing

import (
	"testing"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DefaultTestTimeout defines default test timeout for a testing package
const DefaultTestTimeout = time.Second * 1

// Setup function sets up testing environment. Currently, only uploading schema to the DB.
func Setup() (*ent.Client, error) {
	return db.RunSchemaMigration()
}

// AssertEqualVersion runs assertions on the fields of Version resources.
func AssertEqualVersion(t *testing.T, expected, actual *ent.Version) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.Checksum, actual.Checksum)
}

// AssertDeviceStatus runs assertions on the fields of Device Status resources.
func AssertDeviceStatus(t *testing.T, expected, actual *ent.DeviceStatus) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Status.String(), actual.Status.String())
	assert.Equal(t, expected.LastSeen, actual.LastSeen)
}

// AssertEqualEndpoints runs assertions on the fields of the Endpoint resources.
func AssertEqualEndpoints(t *testing.T, expected, actual *ent.Endpoint) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Host, actual.Host)
	assert.Equal(t, expected.Port, actual.Port)
	assert.Equal(t, expected.Protocol, actual.Protocol)
}

// AssertEqualNetworkDevices runs assertions on the fields of the Network Device resources.
func AssertEqualNetworkDevices(t *testing.T, expected, actual *ent.NetworkDevice) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Model, actual.Model)
	assert.Equal(t, expected.Vendor, actual.Vendor)
	assert.Equal(t, expected.HwVersion, actual.HwVersion)

	require.NotNil(t, expected.Edges.Endpoints)
	require.NotNil(t, actual.Edges.Endpoints)
	assert.Len(t, expected.Edges.Endpoints, len(actual.Edges.Endpoints))

	require.NotNil(t, expected.Edges.SwVersion)
	require.NotNil(t, actual.Edges.SwVersion)
	assert.Equal(t, expected.Edges.SwVersion.Version, actual.Edges.SwVersion.Version)
	assert.Equal(t, expected.Edges.SwVersion.Checksum, actual.Edges.SwVersion.Checksum)

	require.NotNil(t, expected.Edges.FwVersion)
	require.NotNil(t, actual.Edges.FwVersion)
	assert.Equal(t, expected.Edges.FwVersion.Version, actual.Edges.FwVersion.Version)
	assert.Equal(t, expected.Edges.FwVersion.Checksum, actual.Edges.FwVersion.Checksum)
}

// AssertEqualNetworkDevicesNoEdges runs assertions on the non-edges fields of the Network Device resources.
func AssertEqualNetworkDevicesNoEdges(t *testing.T, expected, actual *ent.NetworkDevice) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Model, actual.Model)
	assert.Equal(t, expected.Vendor, actual.Vendor)
	assert.Equal(t, expected.HwVersion, actual.HwVersion)
}

// AssertEqualNetworkDevicesEndpointsOnly runs assertions on the non-edges fields of the Network Device resource
// and on the endpoints edge.
func AssertEqualNetworkDevicesEndpointsOnly(t *testing.T, expected, actual *ent.NetworkDevice) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Model, actual.Model)
	assert.Equal(t, expected.Vendor, actual.Vendor)
	assert.Equal(t, expected.HwVersion, actual.HwVersion)

	require.NotNil(t, expected.Edges.Endpoints)
	require.NotNil(t, actual.Edges.Endpoints)
	assert.Len(t, expected.Edges.Endpoints, len(actual.Edges.Endpoints))

	if len(expected.Edges.Endpoints) == 1 {
		// running assertions on the endpoint
		AssertEqualEndpoints(t, expected.Edges.Endpoints[0], actual.Edges.Endpoints[0])
	}
}
