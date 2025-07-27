// Package server implements main server logic
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/pkg/client/db"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	component     = "component"
	componentName = "grpc-server"
	// server configuration-related constants
	tcpNetwork               = "tcp"
	defaultServerAddress     = "localhost:50051"
	envServerAddress         = "GRPC_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:50051.
	envHTTPServerAddress     = "HTTP_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:80.
	defaultHTTPServerAddress = "localhost:50052"
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

func serve(address string, dbClient *ent.Client, wg *sync.WaitGroup, serverOptions []grpc.ServerOption,
	termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool,
) {
	grpcReadyChan := make(chan bool, 1)
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
			grpcReadyChan <- true
		}
		if err := s.Serve(lis); err != nil {
			zlog.Fatal().Err(err).Msgf("Failed to serve")
		}
	}()

	// starting reverse proxy
	wg.Add(1)
	go func() {
		wg.Add(1) //nolint:staticcheck
		startReverseProxy(address, wg, grpcReadyChan, reverseProxyReadyChan, reverseProxyTermChan)
		wg.Done()
	}()

	// handle termination signals
	termSig := <-termChan
	if termSig {
		zlog.Info().Msg("Gracefully stopping gRPC server")
		s.Stop()
	}
	// report to waitgroup that process is finished
	wg.Done()
}

// startReverseProxy starts the gRPC reverse proxy server which is connected to the HTTP handler.
func startReverseProxy(grpcServerAddress string, wg *sync.WaitGroup, grocReadyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool) {
	// waiting for the gRPC server to start first
	<-grocReadyChan
	zlog.Info().Msg("Starting reverse HTTP proxy")

	// read env variable, where HTTP server is running
	httpServerAddress := os.Getenv(envHTTPServerAddress)
	if httpServerAddress == "" {
		zlog.Warn().Msgf("environment variable \"%s\" is not set, using default address: %s",
			envHTTPServerAddress, defaultHTTPServerAddress)
		httpServerAddress = defaultHTTPServerAddress
	}

	// creating the gRPC-Gateway reverse proxy.
	conn, err := grpc.NewClient(
		grpcServerAddress, // The address of the gRPC server
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to dial to gRPC server")
	}

	mux := runtime.NewServeMux()

	// Registering HTTP handler for our service and connecting the gateway to our gRPC server.
	if err = apiv1.RegisterDeviceMonitoringServiceHandler(context.Background(), mux, conn); err != nil {
		zlog.Fatal().Err(err).Msg("Failed to register HTTP gateway")
	}

	// now, create and start the HTTP server (i.e., our gateway).
	gwServer := &http.Server{
		Addr:    httpServerAddress,
		Handler: mux,
	}

	go func() {
		// On testing will be nil
		if reverseProxyReadyChan != nil {
			reverseProxyReadyChan <- true
		}
		if err = gwServer.ListenAndServe(); err != nil {
			zlog.Fatal().Err(err).Msg("Failed to serve HTTP gateway")
		}
	}()

	// handle termination signals
	termSig := <-reverseProxyTermChan
	if termSig {
		zlog.Info().Msg("Gracefully stopping HTTP server")
		err = gwServer.Shutdown(context.Background())
		if err != nil {
			zlog.Fatal().Err(err).Msg("Failed to gracefully shutdown HTTP gateway")
		}
	}
	// report to waitgroup that process is finished
	wg.Done()
}

// StartServer function configures and brings up gRPC server.
func StartServer(dbClient *ent.Client, wg *sync.WaitGroup, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool) {
	zlog.Info().Msgf("Starting gRPC server...")
	// read env variable, where gRPC server is running
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
	serve(serverAddress, dbClient, wg, serverOptions, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
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
	// DB client doesn't return resource with eager-loaded edges, adding them additionally here
	nd.Edges.Endpoints = endpoints

	// converting to Proto bindings
	protoND := ConvertNetworkDeviceResourceToNetworkDeviceProtoUserSide(nd)
	return &apiv1.AddDeviceResponse{
		Device: protoND,
		Added:  true,
	}, nil
}

func (srv *server) DeleteDevice(ctx context.Context, req *apiv1.DeleteDeviceRequest) (*apiv1.DeleteDeviceResponse, error) {
	zlog.Info().Msgf("Removing network device (%s)", req.GetId())

	resp := &apiv1.DeleteDeviceResponse{
		Id:      req.GetId(),
		Deleted: false,
	}
	err := db.DeleteNetworkDeviceByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		// failed to delete network device
		return resp, err
	}
	// network device was deleted
	resp.Deleted = true
	return resp, nil
}
