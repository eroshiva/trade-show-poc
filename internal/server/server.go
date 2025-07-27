// Package server implements main server logic
package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

const (
	component     = "component"
	componentName = "grpc-server"
	// server configuration-related constants
	tcpNetwork           = "tcp"
	defaultServerAddress = "localhost:50051"
	envServerAddress     = "GRPC_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:50051.
)

var zlog = zerolog.New(zerolog.ConsoleWriter{
	Out:        os.Stderr,
	TimeFormat: time.RFC3339,
	FormatCaller: func(i interface{}) string {
		return filepath.Dir(fmt.Sprintf("%s/", i))
	},
}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

// server implements the DeviceMonitoringServiceServer interface.
type server struct {
	apiv1.DeviceMonitoringServiceServer

	// client for interactions with DB
	dbClient *ent.Client
}

// Options structure defines server's features enablement.
type Options struct {
	EnableInterceptor bool
}

func getServerOptions(_ *Options) ([]grpc.ServerOption, error) {
	// parse server options from configuration
	optionsList := make([]grpc.ServerOption, 0)
	return optionsList, nil
}

func serve(address string, dbClient *ent.Client, serverOptions []grpc.ServerOption, termChan chan bool, readyChan chan bool) {
	lis, err := net.Listen(tcpNetwork, address)
	if err != nil {
		zlog.Fatal().Err(err).Msgf("Failed to listen on %s", address)
	}

	// Create a new gRPC server instance.
	s := grpc.NewServer(serverOptions...)

	gRPCServer := &server{
		dbClient: dbClient,
	}

	// Register our server implementation with the gRPC server.
	apiv1.RegisterDeviceMonitoringServiceServer(s, gRPCServer)

	// Start the server.
	zlog.Info().Msgf("gRPC server listening at %v", lis.Addr())

	go func() {
		// On testing will be nil
		if readyChan != nil {
			readyChan <- true
		}
		if err := s.Serve(lis); err != nil {
			zlog.Fatal().Err(err).Msgf("Failed to serve")
		}
	}()

	// handle termination signals
	termSig := <-termChan
	if termSig {
		s.Stop()
		zlog.Warn().Msg("stopping server")
	}
}

// StartServer function configures and brings up gRPC server.
func StartServer(dbClient *ent.Client, termChan chan bool, readyChan chan bool) {
	zlog.Info().Msgf("Starting server...")
	// read env variable, where server is running
	serverAddress := os.Getenv(envServerAddress)
	if serverAddress == "" {
		zlog.Warn().Msgf("environment variable \"%s\" is not set, using default address: %s",
			envServerAddress, defaultServerAddress)
		serverAddress = defaultServerAddress
	}

	// get server options
	serverOptions, err := getServerOptions(nil)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to get server options")
	}

	// start server
	serve(serverAddress, dbClient, serverOptions, termChan, readyChan)
}

func (srv *server) AddDevice(ctx context.Context, req *apiv1.AddDeviceRequest) (*apiv1.AddDeviceResponse, error) {
	zlog.Info().Msgf("Adding device %s:%s", req.GetDevice().GetVendor(), req.GetDevice().GetModel())

	// if endpoints are set in this request, performing a check if this device already exists
	found := false
	nd := &ent.NetworkDevice{}
	if len(req.GetDevice().GetEndpoints()) > 0 {
		for _, endpoint := range req.GetDevice().GetEndpoints() {
			foundNd, err := db.GetNetworkDeviceByEndpoint(ctx, srv.dbClient, endpoint.GetHost(), endpoint.GetPort())
			if err != nil {
				// network device was not found
				continue
			}
			// device was found
			found = true
			nd = foundNd
			break
		}
	}

	// handling the case when network device already exists
	if found {
		errText := "network device already exists"
		err := fmt.Errorf("%s", errText)
		ndProto := ConvertNetworkDeviceResourceToNetworkDeviceProto(nd)
		// network device already exists, returning error
		return &apiv1.AddDeviceResponse{
			Device:  ndProto,
			Added:   false,
			Details: &errText,
		}, err
	}

	// handling the case, when a network device does not exist.
	// first, creating endpoints.
	endpoints := make([]*ent.Endpoint, 0)
	for _, endpoint := range req.GetDevice().GetEndpoints() {
		protocol := ConvertProtoProtocolToEntProtocol(endpoint.GetProtocol())
		ep, err := db.CreateEndpoint(ctx, srv.dbClient, endpoint.GetHost(), endpoint.GetPort(), protocol)
		if err != nil {
			continue
		}
		endpoints = append(endpoints, ep)
	}

	entVendor := ConvertProtoVendorToEntVendor(req.GetDevice().GetVendor())
	// endpoints are created, now creating network device.
	nd, err := db.CreateNetworkDevice(ctx, srv.dbClient, req.GetDevice().GetModel(), entVendor, endpoints)
	if err != nil {
		return nil, err
	}

	// converting to Proto bindings
	protoND := ConvertNetworkDeviceResourceToNetworkDeviceProto(nd)
	return &apiv1.AddDeviceResponse{
		Device: protoND,
		Added:  true,
	}, nil
}
