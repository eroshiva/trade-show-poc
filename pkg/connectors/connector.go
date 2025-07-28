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
	simulatorv1 "github.com/eroshiva/trade-show-poc/pkg/mocks"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	GetHWVersion(ctx context.Context) (string, error)
	GetSWVersion(ctx context.Context) (*ent.Version, error)
	GetFWVersion(ctx context.Context) (*ent.Version, error)
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

// establishGRPCConnection established gRPC connection with provided endpoint. It returns Network Device
// Simulator client interface for communicating with Device Simulator.
func establishGRPCConnection(ep *ent.Endpoint) (simulatorv1.MockDeviceServiceClient, *grpc.ClientConn, error) {
	serverAddress := CraftServerAddressFromEndpoint(ep)
	// creating the gRPC client
	conn, err := grpc.NewClient(
		serverAddress, // The address of the gRPC server
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to dial to gRPC server %s", serverAddress)
		return nil, nil, err
	}

	return simulatorv1.NewMockDeviceServiceClient(conn), conn, nil
}

// CraftServerAddressFromEndpoint returns string containing server address in the form host:port, e.g., localhost:50051,
// to which connection should be established.
func CraftServerAddressFromEndpoint(ep *ent.Endpoint) string {
	return CraftServerAddress(ep.Host, ep.Port)
}

// CraftServerAddress crafts server address from provided host and port.
func CraftServerAddress(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}
