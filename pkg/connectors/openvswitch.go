// Package connectors implements common interface for all devices and carries individual implementations of each protocol.
package connectors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

const componentNameOpenvSwitch = "openvswitch-connector"

var zlogOVS = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentNameOpenvSwitch).Logger()

// OVSConnector handles status checks for Open vSwitch devices.
type OVSConnector struct {
	Endpoint *ent.Endpoint
}

// GetStatus implements the Connector interface, namely GetStatus function, for Open vSwitch protocol.
func (c *OVSConnector) GetStatus(ctx context.Context) (devicestatus.Status, error) {
	zlogOVS.Info().Msgf("Checking status for %s:%s via Open vSwitch...\n", c.Endpoint.Host, c.Endpoint.Port)
	// Normally, Open vSwitch connection logic goes here, but for now, we'll stick to communication with device simulator.
	client, conn, err := establishGRPCConnection(c.Endpoint)
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Error establishing Open vSwitch connection, reporting that device is down")
		// failed to instantiate connection, reporting device status DOWN.
		return devicestatus.StatusSTATUS_DEVICE_DOWN, nil
	}
	// connection was successfully established, retrieving device status
	resp, err := client.GetStatus(ctx, &emptypb.Empty{})
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to retrieve device status for %s:%s via Open vSwitch", c.Endpoint.Host, c.Endpoint.Port)
		// failed to retrieve device status, returning device status DOWN and an error.
		return devicestatus.StatusSTATUS_DEVICE_DOWN, err
	}
	// device status was successfully retrieved, converting it to correct notation
	status := resp.GetStatus()
	entStatus := server.ConvertProtoStatusToEntStatus(status)
	// gracefully closing connection
	err = conn.Close()
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to gracefully close connection")
	}
	return entStatus, nil
}

// GetHWVersion implements the Connector interface, namely GetHWVersion function, for Open vSwitch protocol.
func (c *OVSConnector) GetHWVersion(ctx context.Context) (string, error) {
	zlogOVS.Info().Msgf("Checking HW version for %s:%s via Open vSwitch...\n", c.Endpoint.Host, c.Endpoint.Port)
	// Normally, Open vSwitch connection logic goes here, but for now, we'll stick to communication with device simulator.
	client, conn, err := establishGRPCConnection(c.Endpoint)
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Error establishing Open vSwitch connection")
		// failed to instantiate connection, returning error
		return "", err
	}
	// connection was successfully established, retrieving device status
	resp, err := client.GetHWVersion(ctx, &emptypb.Empty{})
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to retrieve HW version for %s:%s via Open vSwitch", c.Endpoint.Host, c.Endpoint.Port)
		// failed to retrieve HW version, returning error
		return "", err
	}
	// HW version was successfully retrieved, returning it
	// and gracefully closing connection
	err = conn.Close()
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to gracefully close connection")
	}
	return resp.GetVersion(), nil
}

// GetSWVersion implements the Connector interface, namely GetSWVersion function, for Open vSwitch protocol.
func (c *OVSConnector) GetSWVersion(ctx context.Context) (*ent.Version, error) {
	zlogOVS.Info().Msgf("Checking SW version for %s:%s via Open vSwitch...\n", c.Endpoint.Host, c.Endpoint.Port)
	// Normally, Open vSwitch connection logic goes here, but for now, we'll stick to communication with device simulator.
	client, conn, err := establishGRPCConnection(c.Endpoint)
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Error establishing Open vSwitch connection")
		// failed to instantiate connection, returning error
		return nil, err
	}
	// connection was successfully established, retrieving device status
	resp, err := client.GetSWVersion(ctx, &emptypb.Empty{})
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to retrieve SW version for %s:%s via Open vSwitch", c.Endpoint.Host, c.Endpoint.Port)
		// failed to retrieve SW version, returning error
		return nil, err
	}
	// SW version was successfully retrieved, returning it
	// and gracefully closing connection
	err = conn.Close()
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to gracefully close connection")
	}
	return &ent.Version{
		Version:  resp.GetVersion(),
		Checksum: resp.GetChecksum(),
	}, nil
}

// GetFWVersion implements the Connector interface, namely GetFWVersion function, for Open vSwitch protocol.
func (c *OVSConnector) GetFWVersion(ctx context.Context) (*ent.Version, error) {
	zlogOVS.Info().Msgf("Checking FW version for %s:%s via Open vSwitch...\n", c.Endpoint.Host, c.Endpoint.Port)
	// Normally, Open vSwitch connection logic goes here, but for now, we'll stick to communication with device simulator.
	client, conn, err := establishGRPCConnection(c.Endpoint)
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Error establishing Open vSwitch connection")
		// failed to instantiate connection, returning error
		return nil, err
	}
	// connection was successfully established, retrieving device status
	resp, err := client.GetFWVersion(ctx, &emptypb.Empty{})
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to retrieve FW version for %s:%s via Open vSwitch", c.Endpoint.Host, c.Endpoint.Port)
		// failed to retrieve FW version, returning error
		return nil, err
	}
	// FW version was successfully retrieved, returning it
	// and gracefully closing connection
	err = conn.Close()
	if err != nil {
		zlogOVS.Error().Err(err).Msgf("Failed to gracefully close connection")
	}
	return &ent.Version{
		Version:  resp.GetVersion(),
		Checksum: resp.GetChecksum(),
	}, nil
}
