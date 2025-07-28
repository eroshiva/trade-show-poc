// Package simulatorv1 implements network device simulator means.
package simulatorv1

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	component     = "component"
	componentName = "network-device-simulator"

	// server configuration-related constants
	tcpNetwork = "tcp"
	// EnvServerAddress constant specifies name of the environmental variable for Device Simulator server address.
	EnvServerAddress     = "DEVICE_SIMULATOR_GRPC_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:50151.
	defaultServerAddress = "localhost:50151"

	// EnvDeviceStatus constant specifies name of the device status environmental variable.
	EnvDeviceStatus = "DEVICE_SIMULATOR_DEVICE_STATUS" // Can be "UP", "DOWN", "UNHEALTHY"
	// DeviceStatusUP corresponds to the UP status reported by the device. Default reported value by simulator.
	DeviceStatusUP = "UP"
	// DeviceStatusDOWN corresponds to the DOWN status reported by the device.
	DeviceStatusDOWN = "DOWN"
	// DeviceStatusUNHEALTHY corresponds to the UNHEALTHY status reported by the device.
	DeviceStatusUNHEALTHY = "UNHEALTHY"

	// EnvHWModel constant specifies name of the HW model environmental variable.
	EnvHWModel     = "DEVICE_SIMULATOR_HW_MODEL"
	defaultHWModel = "HW-XYZ"
	// EnvSWVersion constant specifies name of the SW version environmental variable.
	EnvSWVersion     = "DEVICE_SIMULATOR_SW_VERSION"
	defaultSWVersion = "1.0.0"
	// EnvFWVersion constant specifies name of the FW version environmental variable.
	EnvFWVersion     = "DEVICE_SIMULATOR_FW_VERSION"
	defaultFWVersion = "0.1.0"
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// DeviceSimulator is an exportable type of device simulator.
type DeviceSimulator struct {
	simulator *grpc.Server
}

// server implements the MockDeviceServiceServer interface.
type server struct {
	MockDeviceServiceServer
}

func convertDeviceStatus(ds string) apiv1.Status {
	switch ds {
	case DeviceStatusUP:
		return apiv1.Status_STATUS_DEVICE_UP
	case DeviceStatusDOWN:
		return apiv1.Status_STATUS_DEVICE_DOWN
	case DeviceStatusUNHEALTHY:
		return apiv1.Status_STATUS_DEVICE_UNHEALTHY
	default:
		return apiv1.Status_STATUS_UNSPECIFIED
	}
}

// GetStatus returns a status based on the device ID.
func (s *server) GetStatus(_ context.Context, _ *emptypb.Empty) (*apiv1.DeviceStatus, error) {
	zlog.Info().Msg("Received GetStatus request")
	deviceStatus := os.Getenv(EnvDeviceStatus)
	if deviceStatus == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, returning default value: %v",
			EnvDeviceStatus, DeviceStatusUP)
		deviceStatus = DeviceStatusUP
	}
	// value is set, converting and returning it
	status := convertDeviceStatus(deviceStatus)
	if status == apiv1.Status_STATUS_DEVICE_DOWN {
		// returning error
		err := fmt.Errorf("device is unreachable")
		zlog.Info().Msgf("Device status is down, returning an error: %v", err)
		return nil, err
	}

	return &apiv1.DeviceStatus{Status: status}, nil
}

// GetHWVersion returns a mock hardware version.
func (s *server) GetHWVersion(_ context.Context, _ *emptypb.Empty) (*GetVersionResponse, error) {
	zlog.Info().Msgf("Received GetHWVersion request")
	hwModel := os.Getenv(EnvHWModel)
	if hwModel == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default value: %s",
			EnvHWModel, defaultHWModel)
		hwModel = defaultHWModel
	}

	return &GetVersionResponse{Version: hwModel}, nil
}

// GetSWVersion returns a mock software version.
func (s *server) GetSWVersion(_ context.Context, _ *emptypb.Empty) (*apiv1.Version, error) {
	zlog.Info().Msgf("Received GetSWVersion request")
	swVersion := os.Getenv(EnvSWVersion)
	if swVersion == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default value: %s",
			EnvSWVersion, defaultSWVersion)
		swVersion = defaultSWVersion
	}
	fwChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(swVersion)))

	return &apiv1.Version{Version: swVersion, Checksum: fwChecksum}, nil
}

// GetFWVersion returns a mock firmware version.
func (s *server) GetFWVersion(_ context.Context, _ *emptypb.Empty) (*apiv1.Version, error) {
	zlog.Info().Msgf("Received GetFWVersion request")
	fwVersion := os.Getenv(EnvFWVersion)
	if fwVersion == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default value: %s",
			EnvFWVersion, defaultFWVersion)
		fwVersion = defaultFWVersion
	}
	fwChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(fwVersion)))

	return &apiv1.Version{Version: fwVersion, Checksum: fwChecksum}, nil
}

// NewDeviceSimulator is a factory function that creates a network device simulator structure.
func NewDeviceSimulator() *DeviceSimulator {
	return &DeviceSimulator{
		simulator: grpc.NewServer(),
	}
}

// StartNetworkDeviceSimulator function starts network device simulator. Under the hood, it is a pure gRPC server
// implemented for the sake of simplicity and showcasing the interaction.
func (ds *DeviceSimulator) StartNetworkDeviceSimulator() {
	serverAddress := os.Getenv(EnvServerAddress)
	if serverAddress == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default address: %s",
			EnvServerAddress, defaultServerAddress)
		serverAddress = defaultServerAddress
	}

	lis, err := net.Listen(tcpNetwork, serverAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to listen")
	}
	RegisterMockDeviceServiceServer(ds.simulator, &server{})
	go func() {
		zlog.Info().Msgf("gRPC Network Device Simulator listening on %s", serverAddress)
		if err := ds.simulator.Serve(lis); err != nil {
			zlog.Fatal().Err(err).Msgf("failed to serve")
		}
	}()
}

// StopNetworkDeviceSimulator stops gRPC server for network device simulator.
func (ds *DeviceSimulator) StopNetworkDeviceSimulator() {
	zlog.Info().Msg("Gracefully stopping gRPC Network Device Simulator server")
	ds.simulator.Stop()
}
