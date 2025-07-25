// Package db implements utility functions for managing resources in the PostgreSQL.
package db

import (
	"context"
	"fmt"
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
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s", i)) + filepath.Base(fmt.Sprintf("%s", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// CreateNetworkDevice creates a network device resource.
func CreateNetworkDevice(ctx context.Context, client *ent.Client, id, model, hwVersion string, vendor networkdevice.Vendor, endpoints []*ent.Endpoint) (*ent.NetworkDevice, error) {
	u, err := client.NetworkDevice.
		Create().
		SetID(id).
		SetVendor(vendor).
		SetModel(model).
		SetHwVersion(hwVersion).
		AddEndpoints(endpoints...).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create network device")
		return nil, err
	}
	zlog.Debug().Msgf("Creating network device: %v", u)
	return u, nil
}

// GetNetworkDeviceByID retrieves a Network Device resource by ID from the DB.
func GetNetworkDeviceByID(ctx context.Context, client *ent.Client, id string) (*ent.NetworkDevice, error) {
	nd, err := client.NetworkDevice.
		Query().
		Where(networkdevice.ID(id)).
		// `Only` fails if no user found,
		// or more than 1 user returned.
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to get network device")
		return nil, err
	}
	zlog.Debug().Msgf("Network Device has been returned: %v", nd)
	return nd, nil
}

// GetNetworkDeviceByEndpoint retrieves a Network Device resource by the Endpoint resource from the DB.
func GetNetworkDeviceByEndpoint(ctx context.Context, client *ent.Client, host, port string) (*ent.NetworkDevice, error) {
	nd, err := client.NetworkDevice.
		Query().
		Where(networkdevice.HasEndpointsWith(endpoint.Host(host), endpoint.Port(port))).
		// `Only` fails if no user found,
		// or more than 1 user returned.
		Only(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to get network device")
		return nil, err
	}
	zlog.Debug().Msgf("Network Device has been returned: %v", nd)
	return nd, nil
}

// CreateEndpoint creates an Endpoint resource.
func CreateEndpoint(ctx context.Context, client *ent.Client, host, port string, protocol endpoint.Protocol) (*ent.Endpoint, error) {
	ep, err := client.Endpoint.
		Create().
		SetHost(host).
		SetPort(port).
		SetProtocol(protocol).
		Save(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create endpoint")
		return nil, err
	}
	zlog.Debug().Msgf("Creating endpoint: %v", ep)
	return ep, nil
}
