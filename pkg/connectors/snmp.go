// Package connectors implements common interface for all devices and carries individual implementations of each protocol.
package connectors

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/rs/zerolog"
)

const (
	component         = "component"
	componentNameSNMP = "snmp-connector"
)

var (
	zlogSNMP = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			return filepath.Dir(fmt.Sprintf("%s/", i))
		},
	}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentNameSNMP).Logger()

	version    = "XYZ"
	checksum   = fmt.Sprintf("%x", sha256.Sum256([]byte(version)))
	fwVersion  = "fw-" + version
	fwChecksum = fmt.Sprintf("%x", sha256.Sum256([]byte(fwVersion)))
)

// SNMPConnector handles status checks for SNMP devices.
type SNMPConnector struct {
	Endpoint *ent.Endpoint
}

// GetStatus implements the Connector interface, namely GetStatus function, for SNMP protocol.
func (c *SNMPConnector) GetStatus(_ context.Context) (devicestatus.Status, error) {
	zlogSNMP.Info().Msgf("Checking status for %s:%s via SNMP...\n", c.Endpoint.Host, c.Endpoint.Port)
	// SNMP connection logic goes here
	// for now, we'll just report a successful connection.
	return devicestatus.StatusSTATUS_DEVICE_UP, nil
}

// GetHWVersion implements the Connector interface, namely GetHWVersion function, for SNMP protocol.
func (c *SNMPConnector) GetHWVersion(_ context.Context) (string, error) {
	zlogSNMP.Info().Msgf("Checking HW version for %s:%s via SNMP...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded value
	return "hw-XYZ", nil
}

// GetSWVersion implements the Connector interface, namely GetSWVersion function, for SNMP protocol.
func (c *SNMPConnector) GetSWVersion(_ context.Context) (*ent.Version, error) {
	zlogSNMP.Info().Msgf("Checking SW version for %s:%s via SNMP...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded values
	return &ent.Version{
		Version:  version,
		Checksum: checksum,
	}, nil
}

// GetFWVersion implements the Connector interface, namely GetFWVersion function, for SNMP protocol.
func (c *SNMPConnector) GetFWVersion(_ context.Context) (*ent.Version, error) {
	zlogSNMP.Info().Msgf("Checking FW version for %s:%s via SNMP...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded value
	return &ent.Version{
		Version:  fwVersion,
		Checksum: fwChecksum,
	}, nil
}
