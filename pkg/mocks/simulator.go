// Package simulatorv1 implements network device simulator means.
package simulatorv1

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const (
	component     = "component"
	componentName = "network-device-simulator"
	// server configuration-related constants
	tcpNetwork           = "tcp"
	defaultServerAddress = "localhost:50151"
	envServerAddress     = "DEVICE_SIMULATOR_GRPC_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:50151.
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// server implements the MockDeviceServiceServer interface.
type server struct {
	MockDeviceServiceServer
}

// GetStatus returns a status based on the device ID.
func (s *server) GetStatus(_ context.Context, req *GetStatusRequest) (*apiv1.DeviceStatus, error) {
	zlog.Info().Msgf("Received GetStatus request for device %s", req.DeviceId)
	// ToDo - randomly return device status with different probabilities
	// to simulate different device behaviors
	return &apiv1.DeviceStatus{Status: apiv1.Status_STATUS_DEVICE_UP}, nil
}

// GetHWVersion returns a mock hardware version.
func (s *server) GetHWVersion(_ context.Context, req *GetVersionRequest) (*GetVersionResponse, error) {
	// ToDo - read HW version from environmental variable, which is specified on startup
	zlog.Info().Msgf("Received GetHWVersion request for device %s", req.DeviceId)
	return &GetVersionResponse{Version: "HW-XYZ"}, nil
}

// GetSWVersion returns a mock software version.
func (s *server) GetSWVersion(_ context.Context, req *GetVersionRequest) (*apiv1.Version, error) {
	// ToDo - read SW version and its checksum from environmental variable, which is specified on startup
	zlog.Info().Msgf("Received GetSWVersion request for device %s", req.DeviceId)
	return &apiv1.Version{Version: "SW-XYZ", Checksum: "sw-checksum-abc"}, nil
}

// GetFWVersion returns a mock firmware version.
func (s *server) GetFWVersion(_ context.Context, req *GetVersionRequest) (*apiv1.Version, error) {
	// ToDo - read SW version and its checksum from environmental variable, which is specified on startup
	zlog.Info().Msgf("Received GetFWVersion request for device %s", req.DeviceId)
	return &apiv1.Version{Version: "FW-XYZ", Checksum: "fw-checksum-def"}, nil
}

// StartNetworkDeviceSimulator function starts network device simulator. Under the hood, it is a pure gRPC server
// implemented for the sake of simplicity and showcasing the interaction.
func StartNetworkDeviceSimulator() {
	serverAddress := os.Getenv(envServerAddress)
	if serverAddress == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default address: %s",
			envServerAddress, defaultServerAddress)
		serverAddress = defaultServerAddress
	}

	lis, err := net.Listen(tcpNetwork, serverAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to listen")
	}
	s := grpc.NewServer()
	RegisterMockDeviceServiceServer(s, &server{})
	zlog.Info().Msgf("gRPC Network Device Simulator listening on %s", serverAddress)
	if err := s.Serve(lis); err != nil {
		zlog.Fatal().Err(err).Msgf("failed to serve")
	}
}
