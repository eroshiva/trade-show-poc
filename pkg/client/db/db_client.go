// Package db implements utility functions for managing resources in the PostgreSQL.
package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
	"github.com/eroshiva/trade-show-poc/internal/ent/version"
	"github.com/google/uuid"
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
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// CreateNetworkDevice creates a network device resource.
func CreateNetworkDevice(ctx context.Context, client *ent.Client, model string, vendor networkdevice.Vendor, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	// input parameters sanity
	if model == "" || vendor == "" {
		err := fmt.Errorf("model or vendor are unspecified")
		zlog.Error().Err(err).Send()
		return nil, err

	}
	zlog.Debug().Msgf("Creating network device %s by %s with following endpoints: %v", model, vendor, endpoints)

	// generating random ID for the network device
	id := networkDevicePrefix + uuid.NewString()

	nd, err := client.NetworkDevice.
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

	// on create edges are not eager-loaded
	if len(endpoints) > 0 {
		nd.Edges.Endpoints = endpoints
	}
	return nd, nil
}

// UpdateNetworkDeviceByUser is used to update Network Device resource by user. Endpoints are overwritten.
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
	if endpoints != nil || len(endpoints) > 0 {
		nd.Edges.Endpoints = endpoints
	}

	numAfNdNodes, err := client.NetworkDevice.Update().
		Where(networkdevice.ID(id)).
		SetModel(nd.Model).
		SetVendor(nd.Vendor).
		ClearEndpoints(). // cleaning all endpoints out
		AddEndpoints(endpoints...).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	if numAfNdNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of network device didn't return error, number of affected nodes is %d", numAfNdNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return nd, nil
}

// UpdateNetworkDeviceEndpoints is used to update Network Device endpoints by overwriting all endpoints.
func UpdateNetworkDeviceEndpoints(ctx context.Context, client *ent.Client, id string, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Updating network device (%v) endpoints to %v", id, endpoints)
	nd, err := GetNetworkDeviceByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if len(endpoints) == 0 {
		err = errors.New("empty endpoints list")
		zlog.Error().Err(err).Msgf("Please, provide non-empty list of endpoints that you want to update network device(%s) with", id)
		return nil, err
	}

	// overwriting endpoints
	nd.Edges.Endpoints = endpoints
	// adding new endpoints
	numAfNdNodes, err := client.NetworkDevice.Update().
		Where(networkdevice.ID(id)).
		// cleaning all endpoints out
		ClearEndpoints().
		// adding new endpoints
		AddEndpoints(endpoints...).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	if numAfNdNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of network device didn't return error, number of affected nodes is %d", numAfNdNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return nd, nil
}

// UpdateNetworkDeviceAddEndpoints is used to add endpoints to the Network Device by appending endpoints.
func UpdateNetworkDeviceAddEndpoints(ctx context.Context, client *ent.Client, id string, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	zlog.Debug().Msgf("Updating network device (%v) endpoints to %v", id, endpoints)
	nd, err := GetNetworkDeviceByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if len(endpoints) == 0 {
		err = errors.New("empty endpoints list")
		zlog.Error().Err(err).Msgf("Please, provide non-empty list of endpoints that you want to update network device(%s) with", id)
		return nil, err
	}

	// adding endpoints
	nd.Edges.Endpoints = append(nd.Edges.Endpoints, endpoints...)
	numAfNdNodes, err := client.NetworkDevice.Update().
		Where(networkdevice.ID(id)).
		AddEndpoints(endpoints...).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	if numAfNdNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of network device didn't return error, number of affected nodes is %d", numAfNdNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return nd, nil
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
		// check that the Version resource already exists
		retSW, err := GetVersionByVersionAndChecksum(ctx, client, sw.Version, sw.Checksum)
		if err != nil {
			// current SW Version resource does not exist, creating one
			createdSW, err := CreateVersion(ctx, client, sw.Version, sw.Checksum)
			if err != nil {
				// failed to create SW Version resource
				return nil, err
			}
			// SW Version resource created, saving correct resource ID
			sw.ID = createdSW.ID
		}
		if retSW != nil {
			// handling the case when SW Version resource was found, saving correct resource ID
			sw.ID = retSW.ID
		}
		nd.Edges.SwVersion = sw
	}
	if fw != nil {
		// check that the Version resource already exists
		retFW, err := GetVersionByVersionAndChecksum(ctx, client, fw.Version, fw.Checksum)
		if err != nil {
			// current FW Version resource does not exist, creating one
			createdFW, err := CreateVersion(ctx, client, fw.Version, fw.Checksum)
			if err != nil {
				// failed to create FW Version resource
				return nil, err
			}
			// FW Version resource created, saving correct resource ID
			fw.ID = createdFW.ID
		}
		if retFW != nil {
			// handling the case when FW Version resource was found, saving correct resource ID
			fw.ID = retFW.ID
		}
		nd.Edges.FwVersion = fw
	}

	numAfNdNodes, err := client.NetworkDevice.Update().
		Where(networkdevice.ID(id)).
		SetHwVersion(nd.HwVersion).
		SetSwVersion(nd.Edges.SwVersion).
		SetFwVersion(nd.Edges.FwVersion).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update network device (%s)", id)
		return nil, err
	}

	if numAfNdNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of network device didn't return error, number of affected nodes is %d", numAfNdNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return nd, nil
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

// ListNetworkDevices retrieves all Network Device resources present in the system.
func ListNetworkDevices(ctx context.Context, client *ent.Client) ([]*ent.NetworkDevice, error) {
	zlog.Debug().Msg("Listing all network devices")
	nds, err := client.NetworkDevice.Query().
		// Eager-loading edges
		WithEndpoints().
		WithSwVersion().
		WithFwVersion().
		All(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to list network devices")
		return nil, err
	}

	return nds, nil
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
	// input parameters sanity
	if host == "" || port == "" || protocol == "" {
		err := fmt.Errorf("one of the input parameters is missing")
		zlog.Error().Err(err).Msgf("Failed to create %s endpoint on %s:%s", protocol, host, port)
		return nil, err
	}
	zlog.Debug().Msgf("Creating %s endpoint on %s:%s", protocol, host, port)

	// generating random ID for the network device
	id := endpointPrefix + uuid.NewString()
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

	// setting fields
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
	numAfEpNodes, err := client.Endpoint.Update().
		Where(endpoint.ID(id)).
		SetHost(ep.Host).
		SetPort(ep.Port).
		SetProtocol(ep.Protocol).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update endpoint (%s)", id)
		return nil, err
	}

	if numAfEpNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of endpoint didn't return error, number of affected nodes is %d", numAfEpNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return ep, nil
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

// ListEndpoints lists all endpoint resources present in the system.
func ListEndpoints(ctx context.Context, client *ent.Client) ([]*ent.Endpoint, error) {
	zlog.Debug().Msgf("Listing endpoints")
	eps, err := client.Endpoint.Query().All(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to list endpoints")
		return nil, err
	}

	return eps, nil
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
func CreateDeviceStatus(ctx context.Context, client *ent.Client, status devicestatus.Status, lastSeen string, cal int32, nd *ent.NetworkDevice) (*ent.DeviceStatus, error) {
	// input parameters sanity
	if nd == nil {
		err := fmt.Errorf("network device resource should be specified")
		zlog.Error().Err(err).Msg("Failed to create device status")
		return nil, err
	}
	if status == "" {
		err := fmt.Errorf("status must be specified")
		zlog.Error().Err(err).Msg("Failed to create device status")
		return nil, err
	}

	zlog.Debug().Msgf("Creating device status (%s at %s) for network device (%s)", status, lastSeen, nd.ID)
	// creating device status ID
	id := deviceStatusPrefix + uuid.NewString()
	ds, err := client.DeviceStatus.Create().
		SetID(id).
		SetStatus(status).
		SetLastSeen(lastSeen).
		SetConsequentialFailedConnectivityAttempts(cal).
		SetNetworkDevice(nd).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create device status")
		return nil, err
	}

	return ds, nil
}

// GetDeviceStatusByID retrieves device status resource by provided ID.
func GetDeviceStatusByID(ctx context.Context, client *ent.Client, id string) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving device status (%s)", id)

	ds, err := client.DeviceStatus.Query().Where(devicestatus.ID(id)).Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get device status (%s)", id)
		return nil, err
	}

	return ds, nil
}

// ListDeviceStatuses retrieves all device statuses present in the DB.
func ListDeviceStatuses(ctx context.Context, client *ent.Client) ([]*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving all device statuses")

	dss, err := client.DeviceStatus.Query().
		// Eager-loading network device resources
		WithNetworkDevice().
		All(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get all device statuses")
		return nil, err
	}

	return dss, nil
}

// GetDeviceStatusByNetworkDeviceID retrieves device status resource by provided network device ID.
func GetDeviceStatusByNetworkDeviceID(ctx context.Context, client *ent.Client, networkDeviceID string) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving device status by network device (%s)", networkDeviceID)

	ds, err := client.DeviceStatus.Query().
		Where(devicestatus.HasNetworkDeviceWith(networkdevice.ID(networkDeviceID))).
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get device status by network device (%s)", networkDeviceID)
		return nil, err
	}

	return ds, nil
}

// GetDeviceStatusByEndpointID retrieves device status resource by provided endpoint ID.
func GetDeviceStatusByEndpointID(ctx context.Context, client *ent.Client, endpointID string) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving device status by endpoint (%s)", endpointID)

	ds, err := client.DeviceStatus.Query().
		Where(devicestatus.HasNetworkDeviceWith(networkdevice.HasEndpointsWith(endpoint.ID(endpointID)))).
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get device status by endpoint (%s)", endpointID)
		return nil, err
	}

	return ds, nil
}

// ListDeviceStatusResources lists all device status resources available in the DB.
func ListDeviceStatusResources(ctx context.Context, client *ent.Client) ([]*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Retrieving all device status resources from the DB")
	dss, err := client.DeviceStatus.Query().All(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to retrieve all device status resources")
		return nil, err
	}

	return dss, nil
}

// UpdateDeviceStatusByNetworkDeviceID updates device status for the network device with provided ID. If device status for this
// network device does not exist, it creates one.
func UpdateDeviceStatusByNetworkDeviceID(ctx context.Context, client *ent.Client, networkDeviceID string, status devicestatus.Status, lastSeen string, cal int32) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Updating device status resource by network device (%s)", networkDeviceID)
	ds, err := GetDeviceStatusByNetworkDeviceID(ctx, client, networkDeviceID)
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
		ds, err := CreateDeviceStatus(ctx, client, status, lastSeen, cal, nd)
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
	numAfDsNodes, err := client.DeviceStatus.Update().
		Where(devicestatus.ID(ds.ID)).
		SetStatus(ds.Status).
		SetLastSeen(ds.LastSeen).
		SetConsequentialFailedConnectivityAttempts(cal).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update device status for network device (%s)", networkDeviceID)
		return nil, err
	}
	if numAfDsNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of device status didn't return error, number of affected nodes is %d", numAfDsNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return ds, nil
}

// UpdateDeviceStatusByEndpointID updates device status for the network device with existing endpoint with provided ID. If device status for this
// endpoint and network device does not exist, it creates one.
func UpdateDeviceStatusByEndpointID(ctx context.Context, client *ent.Client, endpointID string, status devicestatus.Status, lastSeen string, cal int32) (*ent.DeviceStatus, error) {
	zlog.Debug().Msgf("Updating device status resource by endpoint (%s)", endpointID)
	ds, err := GetDeviceStatusByEndpointID(ctx, client, endpointID)
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
		ds, err := CreateDeviceStatus(ctx, client, status, lastSeen, cal, nd)
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

	numAfDsNodes, err := client.DeviceStatus.Update().
		Where(devicestatus.ID(ds.ID)).
		SetStatus(ds.Status).
		SetLastSeen(ds.LastSeen).
		SetConsequentialFailedConnectivityAttempts(cal).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update device status for endpoint (%s)", endpointID)
		return nil, err
	}
	if numAfDsNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of device status didn't return error, number of affected nodes is %d", numAfDsNodes)
		zlog.Error().Err(newErr).Send()
		return nil, err
	}

	return ds, nil
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
	// input parameters sanity
	if version == "" || checksum == "" {
		err := fmt.Errorf("version or checksum is unspecified")
		zlog.Error().Err(err).Msgf("Failed to create version resource")
		return nil, err
	}
	zlog.Debug().Msgf("Creating version resource (%s:%s)", version, checksum)
	// creating resource ID
	id := versionPrefix + uuid.NewString()
	v, err := client.Version.Create().SetID(id).SetVersion(version).SetChecksum(checksum).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to create version resource(%s:%s)", version, checksum)
		return nil, err
	}

	return v, nil
}

// ListVersions lists all version resources available in the DB.
func ListVersions(ctx context.Context, client *ent.Client) ([]*ent.Version, error) {
	zlog.Debug().Msgf("Retrieving all version resources from the DB")
	v, err := client.Version.Query().All(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to retrieve all version resources")
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

// GetVersionByVersionAndChecksum retrieves version resource by provided version and checksum.
func GetVersionByVersionAndChecksum(ctx context.Context, client *ent.Client, vrs, checksum string) (*ent.Version, error) {
	zlog.Debug().Msgf("Retrieving version resource (%s:%s)", vrs, checksum)
	v, err := client.Version.Query().Where(version.Version(vrs), version.Checksum(checksum)).Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to get version resource (%s:%s)", vrs, checksum)
		return nil, err
	}

	return v, nil
}

// UpdateVersion updates version resource fields by provided resource ID.
func UpdateVersion(ctx context.Context, client *ent.Client, id, vers, checksum string) (*ent.Version, error) {
	zlog.Debug().Msgf("Updating version resource (%s)", id)
	v, err := GetVersionByID(ctx, client, id)
	if err != nil {
		return nil, err
	}

	if vers != "" {
		v.Version = vers
	}
	if checksum != "" {
		v.Checksum = checksum
	}
	numAfVNodes, err := client.Version.Update().Where(version.ID(id)).SetVersion(v.Version).SetChecksum(v.Checksum).Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to update version resource (%s)", id)
		return nil, err
	}
	if numAfVNodes != 1 {
		// something bad has happened, returning error
		newErr := fmt.Errorf("update of version didn't return error, number of affected nodes is %d", numAfVNodes)
		zlog.Error().Err(newErr).Send()
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
