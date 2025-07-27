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

// GetStatus implements the Connector interface for RESTCONF protocol.
func (c *RESTCONFConnector) GetStatus(_ context.Context) (devicestatus.Status, error) {
	zlogRESTCONF.Info().Msgf("Checking status for %s:%s via RESTCONF...\n", c.Endpoint.Host, c.Endpoint.Port)
	// RESTCONF connection logic goes here
	// For this example, we'll just report a successful connection.
	return devicestatus.StatusSTATUS_DEVICE_UP, nil
}
