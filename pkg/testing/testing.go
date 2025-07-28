// Package testing contains helper functions to run unit tests within this repository.
package testing

import (
	"sync"
	"testing"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DefaultTestTimeout defines default test timeout for a testing package
const (
	DefaultTestTimeout           = time.Second * 1
	defaultGRPCTestServerAddress = "localhost:50051"
	defaultHTTPTestServerAddress = "localhost:50052"
)

// Setup function sets up testing environment. Currently, only uploading schema to the DB.
func Setup() (*ent.Client, error) {
	return db.RunSchemaMigration()
}

// SetupFull function sets up testing environment.It uploads schema to the DB and starts gRPC and HTTP reverse proxy servers.
func SetupFull(grpcServerAddress, httpServerAddress string) (*ent.Client, apiv1.DeviceMonitoringServiceClient, *sync.WaitGroup, chan bool, chan bool, error) {
	if grpcServerAddress == "" {
		grpcServerAddress = defaultGRPCTestServerAddress
	}

	if httpServerAddress == "" {
		httpServerAddress = defaultHTTPTestServerAddress
	}

	client, err := db.RunSchemaMigration()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	wg := &sync.WaitGroup{}
	termChan := make(chan bool, 1)
	readyChan := make(chan bool, 1)
	reverseProxyReadyChan := make(chan bool, 1)
	reverseProxyTermChan := make(chan bool, 1)
	wg.Add(1)
	go func() {
		wg.Add(1) //nolint:staticcheck
		server.StartServer(grpcServerAddress, httpServerAddress, client, wg, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
		wg.Done()
	}()
	// Waiting until both servers are up and running
	<-readyChan
	<-reverseProxyReadyChan

	// creating gRPC testing client
	conn, err := grpc.NewClient(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	grpcClient := apiv1.NewDeviceMonitoringServiceClient(conn)

	return client, grpcClient, wg, termChan, reverseProxyTermChan, nil
}

// TeardownFull function tears down testing suite including DB connection, gRPC and HTTP reverse proxy servers.
func TeardownFull(client *ent.Client, wg *sync.WaitGroup, termChan, reverseProxyTermChan chan bool) {
	close(termChan)
	close(reverseProxyTermChan)
	err := db.GracefullyCloseDBClient(client)
	if err != nil {
		panic(err)
	}
	wg.Wait()
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
