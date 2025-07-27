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
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/rs/zerolog"
)

const componentName = "connector"

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// Connector defines the interface for connecting to a network device
// and retrieving its status.
type Connector interface {
	GetStatus(ctx context.Context) (devicestatus.Status, error)
}

// NewConnector function returns the correct connector for a given endpoint protocol.
func NewConnector(ep *ent.Endpoint) (Connector, error) {
	switch ep.Protocol {
	case endpoint.ProtocolPROTOCOL_SNMP:
		return &SNMPConnector{Endpoint: ep}, nil
	case endpoint.ProtocolPROTOCOL_RESTCONF:
		return &RESTCONFConnector{Endpoint: ep}, nil
	case endpoint.ProtocolPROTOCOL_NETCONF:
		return &NETCONFConnector{Endpoint: ep}, nil
	case endpoint.ProtocolPROTOCOL_OPEN_V_SWITCH:
		return &OVSConnector{Endpoint: ep}, nil
	default:
		err := fmt.Errorf("unsupported protocol: %s", ep.Protocol)
		zlog.Warn().Err(err).Msgf("Protocol %s is not supported", ep.Protocol)
		return nil, err
	}
}
