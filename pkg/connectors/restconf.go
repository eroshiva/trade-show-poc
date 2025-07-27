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

const componentNameRESTCONF = "restconf-connector"

var zlogRESTCONF = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentNameRESTCONF).Logger()

// RESTCONFConnector handles status checks for RESTCONF devices.
type RESTCONFConnector struct {
	Endpoint *ent.Endpoint
}

// GetStatus implements the Connector interface, namely GetStatus function, for RESTCONF protocol.
func (c *RESTCONFConnector) GetStatus(_ context.Context) (devicestatus.Status, error) {
	zlogRESTCONF.Info().Msgf("Checking status for %s:%s via RESTCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// RESTCONF connection logic goes here
	// for now, we'll just report a successful connection.
	return devicestatus.StatusSTATUS_DEVICE_UP, nil
}

// GetHWVersion implements the Connector interface, namely GetHWVersion function, for RESTCONF protocol.
func (c *RESTCONFConnector) GetHWVersion(_ context.Context) (string, error) {
	zlogRESTCONF.Info().Msgf("Checking HW version for %s:%s via RESTCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded value
	return "hw-XYZ", nil
}

// GetSWVersion implements the Connector interface, namely GetSWVersion function, for RESTCONF protocol.
func (c *RESTCONFConnector) GetSWVersion(_ context.Context) (*ent.Version, error) {
	zlogRESTCONF.Info().Msgf("Checking SW version for %s:%s via RESTCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded values
	return &ent.Version{
		Version:  version,
		Checksum: checksum,
	}, nil
}

// GetFWVersion implements the Connector interface, namely GetFWVersion function, for RESTCONF protocol.
func (c *RESTCONFConnector) GetFWVersion(_ context.Context) (*ent.Version, error) {
	zlogRESTCONF.Info().Msgf("Checking FW version for %s:%s via RESTCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// for now, we just return hardcoded value
	return &ent.Version{
		Version:  fwVersion,
		Checksum: fwChecksum,
	}, nil
}
