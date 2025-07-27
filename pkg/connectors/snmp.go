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

const (
	component         = "component"
	componentNameSNMP = "snmp-connector"
)

var zlogSNMP = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentNameSNMP).Logger()

// SNMPConnector handles status checks for SNMP devices.
type SNMPConnector struct {
	Endpoint *ent.Endpoint
}

// GetStatus implements the Connector interface for SNMP protocol.
func (c *SNMPConnector) GetStatus(_ context.Context) (devicestatus.Status, error) {
	zlogSNMP.Info().Msgf("Checking status for %s:%s via SNMP...\n", c.Endpoint.Host, c.Endpoint.Port)
	// SNMP connection logic goes here
	// For this example, we'll just report a successful connection.
	return devicestatus.StatusSTATUS_DEVICE_UP, nil
}
