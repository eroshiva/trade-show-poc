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
	"github.com/rs/zerolog"
)

const componentNameNETCONF = "netconf-connector"

var zlogNETCONF = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentNameNETCONF).Logger()

// NETCONFConnector handles status checks for NETCONF devices.
type NETCONFConnector struct {
	Endpoint *ent.Endpoint
}

// GetStatus implements the Connector interface, namely GetStatus function, for NETCONF protocol.
func (c *NETCONFConnector) GetStatus(_ context.Context) (devicestatus.Status, error) {
	zlogNETCONF.Info().Msgf("Checking status for %s:%s via NETCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// NETCONF connection logic goes here
	// for now, we'll just report a successful connection.
	return devicestatus.StatusSTATUS_DEVICE_UP, nil
}

// GetHWVersion implements the Connector interface, namely GetHWVersion function, for NETCONF protocol.
func (c *NETCONFConnector) GetHWVersion(_ context.Context) (string, error) {
	zlogNETCONF.Info().Msgf("Checking HW version for %s:%s via NETCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded value
	return "hw-XYZ", nil
}

// GetSWVersion implements the Connector interface, namely GetSWVersion function, for NETCONF protocol.
func (c *NETCONFConnector) GetSWVersion(_ context.Context) (*ent.Version, error) {
	zlogNETCONF.Info().Msgf("Checking SW version for %s:%s via NETCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded values
	return &ent.Version{
		Version:  version,
		Checksum: checksum,
	}, nil
}

// GetFWVersion implements the Connector interface, namely GetFWVersion function, for NETCONF protocol.
func (c *NETCONFConnector) GetFWVersion(_ context.Context) (*ent.Version, error) {
	zlogNETCONF.Info().Msgf("Checking FW version for %s:%s via NETCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded value
	return &ent.Version{
		Version:  fwVersion,
		Checksum: fwChecksum,
	}, nil
}
