// Package db implements utility functions for managing resources in the PostgreSQL.
package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/ent/version"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
	"github.com/rs/zerolog"
)

const (
	component     = "component"
	componentName = "db-client"
	// defining prefixes of the resources for the sake of easy post-manipulation
	networkDevicePrefix = "netdev-"
	endpointPrefix      = "endpoint-"
	deviceStatusPrefix  = "devstat-"
	versionPrefix       = "version-"
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s", i)) + filepath.Base(fmt.Sprintf("%s", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// CreateNetworkDevice creates a network device resource.
func CreateNetworkDevice(ctx context.Context, client *ent.Client, model string, vendor networkdevice.Vendor, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Creating network device %s by %s with following endpoints: %v", vendor, model, endpoints)

	// generating random ID for the network device
	id := networkDevicePrefix + uuid.New().String()

	u, err := client.NetworkDevice.
		Create().
		SetID(id).
		SetVendor(vendor).
		SetModel(model).
		AddEndpoints(endpoints...).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create new network device")
		return nil, err
	}

	return u, nil
}

// UpdateNetworkDeviceByUser is used to update Network Device resource by user.
func UpdateNetworkDeviceByUser(ctx context.Context, client *ent.Client, id, model string, vendor networkdevice.Vendor, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Updating network device (%v)", id)
	nd, err := GetNetworkDeviceByID(ctx, client, id)
	if err != nil {
		return nil, err
	}
	if model != "" {
		nd.Model = model
	}
	if vendor != "" {
		nd.Vendor = vendor
	}
	if endpoints != nil {
		nd.Edges.Endpoints = endpoints
	}

	updNd, err := client.NetworkDevice.UpdateOne(nd).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	return updNd, nil
}

// UpdateNetworkDeviceEndpoints is used to update Network Device endpoints (by user).
func UpdateNetworkDeviceEndpoints(ctx context.Context, client *ent.Client, id string, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Updating network device (%v) endpoints to %v", id, endpoints)
	nd, err := GetNetworkDeviceByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if endpoints == nil {
		err = errors.New("empty endpoints list")
		zlog.Error().Err(err).Msgf("Please, provide non-empty list of endpoints that you want to update network device(%s) with", id)
		return nil, err
	}

	// overwriting endpoints
	nd.Edges.Endpoints = endpoints

	updNd, err := client.NetworkDevice.UpdateOne(nd).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	return updNd, nil
}

// UpdateNetworkDeviceAddEndpoints is used to add endpoints to the Network Device (by user).
func UpdateNetworkDeviceAddEndpoints(ctx context.Context, client *ent.Client, id string, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Updating network device (%v) endpoints to %v", id, endpoints)
	nd, err := GetNetworkDeviceByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if endpoints == nil {
		err = errors.New("empty endpoints list")
		zlog.Error().Err(err).Msgf("Please, provide non-empty list of endpoints that you want to update network device(%s) with", id)
		return nil, err
	}

	// overwriting endpoints
	nd.Edges.Endpoints = append(nd.Edges.Endpoints, endpoints...)
	updNd, err := client.NetworkDevice.UpdateOne(nd).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	return updNd, nil
}

// UpdateNetworkDeviceVersions is used to update Network Device HW, SW, and FW versions.
func UpdateNetworkDeviceVersions(ctx context.Context, client *ent.Client, id, hw string, sw, fw *ent.Version) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Updating network device (%v) HW (%s), SW (%v), and FW (%v) versions", id, hw, sw, fw)
	nd, err := GetNetworkDeviceByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if hw != "" {
		nd.HwVersion = hw
	}
	if sw != nil {
		nd.Edges.SwVersion = sw
	}
	if fw != nil {
		nd.Edges.FwVersion = fw
	}

	updNd, err := client.NetworkDevice.UpdateOne(nd).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	return updNd, nil
}

// GetNetworkDeviceByID retrieves a Network Device resource by ID from the DB.
func GetNetworkDeviceByID(ctx context.Context, client *ent.Client, id string) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Retrieving network device by ID: %s", id)
	nd, err := client.NetworkDevice.
		Query().
		Where(networkdevice.ID(id)).
		// Eager-loading all edges
		WithEndpoints().
		WithFwVersion().
		WithSwVersion().
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get network device (%s)", id)
		return nil, err
	}

	return nd, nil
}

// GetNetworkDeviceByEndpoint retrieves a Network Device resource by the Endpoint resource from the DB.
func GetNetworkDeviceByEndpoint(ctx context.Context, client *ent.Client, host, port string) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Retrieving network device by endpoint %s:%s", host, port)
	nd, err := client.NetworkDevice.
		Query().
		Where(networkdevice.HasEndpoints()).
		Where(networkdevice.HasEndpointsWith(endpoint.Host(host), endpoint.Port(port))).
		// Eager-loading all edges
		WithFwVersion().
		WithSwVersion().
		WithEndpoints().
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get network device by host %s:%s", host, port)
		return nil, err
	}

	return nd, nil
}

// DeleteNetworkDeviceByID deletes network device by provided ID.
func DeleteNetworkDeviceByID(ctx context.Context, client *ent.Client, id string) error {
	zlog.Debug().Msgf("Deleting network device (%s)", id)
	_, err := client.NetworkDevice.Delete().Where(networkdevice.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to delete network device (%s)", id)
		return err
	}

	return nil
}

// CreateEndpoint creates an Endpoint resource.
func CreateEndpoint(ctx context.Context, client *ent.Client, host, port string, protocol endpoint.Protocol) (*ent.Endpoint, error) {
	zlog.Debug().Msgf("Creating %s endpoint on %s:%s", protocol, host, port)

	// generating random ID for the network device
	id := endpointPrefix + uuid.New().String()
	ep, err := client.Endpoint.
		Create().
		SetID(id).
		SetHost(host).
		SetPort(port).
		SetProtocol(protocol).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create endpoint")
		return nil, err
	}

	return ep, nil
}

// UpdateEndpoint updates endpoint associated with provided ID with new details.
func UpdateEndpoint(ctx context.Context, client *ent.Client, id string, host, port string, protocol endpoint.Protocol) (*ent.Endpoint, error) {
	zlog.Debug().Msgf("Updating %s endpoint (%s) on %s:%s", protocol, id, host, port)
	// retrieving endpoint
	ep, err := GetEndpointByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	// setting fileds
	if host != "" {
		ep.Host = host
	}
	if port != "" {
		ep.Port = port
	}
	if protocol != "" {
		ep.Protocol = protocol
	}

	// updating endpoint
	updEp, err := client.Endpoint.UpdateOne(ep).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update endpoint (%s)", id)
		return nil, err
	}

	return updEp, nil
}

// GetEndpointByID retrieves endpoint by ID.
func GetEndpointByID(ctx context.Context, client *ent.Client, id string) (*ent.Endpoint, error) {
	zlog.Debug().Msgf("Retrieving endpoint (%s)", id)
	ep, err := client.Endpoint.Query().
		Where(endpoint.ID(id)).
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get endpoint (%s)", id)
		return nil, err
	}

	return ep, nil
}

// DeleteEndpointByID deletes endpoint resource by provided ID.
func DeleteEndpointByID(ctx context.Context, client *ent.Client, id string) error {
	zlog.Debug().Msgf("Deleting endpoint (%s)", id)
	_, err := client.Endpoint.Delete().Where(endpoint.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to delete endpoint (%s)", id)
		return err
	}

	return nil
}

// CreateDeviceStatus creates a device status resource for the specific network device.
func CreateDeviceStatus(ctx context.Context, client *ent.Client, status devicestatus.Status, lastSeen string, nd *ent.NetworkDevice) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Creating device status (%s at %s) for network device (%s)", status, lastSeen, nd.ID)

	// creating device status ID
	id := deviceStatusPrefix + uuid.New().String()
	ds, err := client.DeviceStatus.Create().
		SetID(id).
		SetStatus(status).
		SetLastSeen(lastSeen).
		SetNetworkDevice(nd).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create device status")
		return nil, err
	}

	return ds, nil
}

// UpdateDeviceStatusByNetworkDeviceID updates device status for the network device with provided ID. If device status for this
// network device does not exist, it creates one.
func UpdateDeviceStatusByNetworkDeviceID(ctx context.Context, client *ent.Client, networkDeviceID string, status devicestatus.Status, lastSeen string) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving device status resource for network device (%s)", networkDeviceID)
	ds, err := client.DeviceStatus.Query().
		Where(devicestatus.HasNetworkDeviceWith(networkdevice.ID(networkDeviceID))).
		Only(ctx)
	if err != nil {
		// no device status for a given network device has been found, creating one
		zlog.Error().Err(err).Msgf("Failed to get device status for network device (%s)", networkDeviceID)
		// retrieving a network device first
		nd, err := GetNetworkDeviceByID(ctx, client, networkDeviceID)
		if err != nil {
			// no network device resource exists yet, creation should start from here
			zlog.Error().Err(err).Msgf("Failed to get network device (%s) - it should be created first in the system", networkDeviceID)
			return nil, err
		}
		// creating device status
		ds, err := CreateDeviceStatus(ctx, client, status, lastSeen, nd)
		if err != nil {
			return nil, err
		}
		return ds, nil
	}
	// device status was found, updating it
	if status != "" {
		ds.Status = status
	}
	if lastSeen != "" {
		ds.LastSeen = lastSeen
	}
	// updating device status in the DB.
	updDs, err := client.DeviceStatus.UpdateOne(ds).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update device status for network device (%s)", networkDeviceID)
		return nil, err
	}

	return updDs, nil
}

// UpdateDeviceStatusByEndpointID updates device status for the network device with existing endpoint with provided ID. If device status for this
// endpoint and network device does not exist, it creates one.
func UpdateDeviceStatusByEndpointID(ctx context.Context, client *ent.Client, endpointID string, status devicestatus.Status, lastSeen string) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving device status resource for endpoint (%s)", endpointID)
	ds, err := client.DeviceStatus.Query().
		Where(devicestatus.HasNetworkDeviceWith(networkdevice.HasEndpointsWith(endpoint.ID(endpointID)))).
		Only(ctx)
	if err != nil {
		// no device status was found for this endpoint, creating one
		zlog.Error().Err(err).Msgf("Failed to get device status for endpoint (%s)", endpointID)
		// retrieving endpoint by ID
		ep, err := GetEndpointByID(ctx, client, endpointID)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to get endpoint (%s) - it should be created in the system first", endpointID)
			return nil, err
		}
		// retrieving a network device by endpoint ID
		nd, err := GetNetworkDeviceByEndpoint(ctx, client, ep.Host, ep.Port)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to get network device by endpoint (%s) - it should be created first in the system", endpointID)
			return nil, err
		}
		// creating device status
		ds, err := CreateDeviceStatus(ctx, client, status, lastSeen, nd)
		if err != nil {
			return nil, err
		}
		return ds, nil
	}

	if status != "" {
		ds.Status = status
	}
	if lastSeen != "" {
		ds.LastSeen = lastSeen
	}
	updDs, err := client.DeviceStatus.UpdateOne(ds).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update device status for endpoint (%s)", endpointID)
		return nil, err
	}

	return updDs, nil
}

// DeleteDeviceStatusByID deletes device status resource by provided ID.
func DeleteDeviceStatusByID(ctx context.Context, client *ent.Client, id string) error {
	zlog.Debug().Msgf("Deleting device status (%s)", id)
	_, err := client.DeviceStatus.Delete().Where(devicestatus.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to delete device status (%s)", id)
		return err
	}

	return nil
}

// CreateVersion creates a version resource in the DB.
func CreateVersion(ctx context.Context, client *ent.Client, version, checksum string) (*ent.Version, error) {
	zlog.Debug().Msgf("Creating version resource (%s:%s)", version, checksum)
	// creating resource ID
	id := versionPrefix + uuid.New().String()
	v, err := client.Version.Create().SetID(id).SetVersion(version).SetChecksum(checksum).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to create version resource(%s:%s)", version, checksum)
		return nil, err
	}

	return v, nil
}

// GetVersionByID retrieves version resource by provided resource ID.
func GetVersionByID(ctx context.Context, client *ent.Client, id string) (*ent.Version, error) {
	zlog.Debug().Msgf("Retrieving version resource (%s)", id)
	v, err := client.Version.Query().Where(version.ID(id)).Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get version resource (%s)", id)
		return nil, err
	}

	return v, nil
}

// UpdateVersion updates version resource fields by provided resource ID.
func UpdateVersion(ctx context.Context, client *ent.Client, id, version, checksum string) (*ent.Version, error) {
	zlog.Debug().Msgf("Updating version resource (%s)", id)
	v, err := GetVersionByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if version != "" {
		v.Version = version
	}
	if checksum != "" {
		v.Checksum = checksum
	}
	_, err = client.Version.UpdateOne(v).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update version resource (%s)", id)
		return nil, err
	}

	return v, nil
}

// DeleteVersionByID deletes version resource with provided ID.
func DeleteVersionByID(ctx context.Context, client *ent.Client, id string) error {
	zlog.Debug().Msgf("Deleting version resource (%s)", id)
	_, err := client.Version.Delete().Where(version.ID(id)).Exec(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to delete version resource (%s)", id)
		return err
	}

	return nil
}
