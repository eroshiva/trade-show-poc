// Package server implements main server logic
package server

import (
	"context"
	"errors"
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
	"google.golang.org/protobuf/types/known/emptypb"
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

func serve(grpcAddress, httpAddress string, dbClient *ent.Client, wg *sync.WaitGroup, serverOptions []grpc.ServerOption,
	termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool,
) {
	grpcReadyChan := make(chan bool, 1)
	lis, err := net.Listen(tcpNetwork, grpcAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msgf("Failed to listen on %s", grpcAddress)
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
		startReverseProxy(grpcAddress, httpAddress, wg, grpcReadyChan, reverseProxyReadyChan, reverseProxyTermChan)
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
func startReverseProxy(grpcServerAddress, httpServerAddress string, wg *sync.WaitGroup, grocReadyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool) {
	// waiting for the gRPC server to start first
	<-grocReadyChan
	zlog.Info().Msg("Starting reverse HTTP proxy")

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

// GetGRPCServerAddress function reads environmental variable and returns a gRPC server address.
func GetGRPCServerAddress() string {
	// read env variable, where gRPC server is running
	serverAddress := os.Getenv(envServerAddress)
	if serverAddress == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default gRPC server address: %s",
			envServerAddress, defaultServerAddress)
		serverAddress = defaultServerAddress
	}
	return serverAddress
}

// GetHTTPServerAddress function reads environmental variable and returns an HTTP server address.
func GetHTTPServerAddress() string {
	// read env variable, where HTTP server is running
	httpServerAddress := os.Getenv(envHTTPServerAddress)
	if httpServerAddress == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default address: %s",
			envHTTPServerAddress, defaultHTTPServerAddress)
		httpServerAddress = defaultHTTPServerAddress
	}
	return httpServerAddress
}

// StartServer function configures and brings up gRPC server.
func StartServer(gRPCServerAddress, httpServerAddress string, dbClient *ent.Client, wg *sync.WaitGroup, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool) {
	zlog.Info().Msgf("Starting gRPC server...")

	// get server options
	serverOptions, err := getServerOptions(nil)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to get server options")
	}

	// start server
	serve(gRPCServerAddress, httpServerAddress, dbClient, wg, serverOptions, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
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
	protoND := ConvertNetworkDeviceResourceToNetworkDeviceProto(nd)
	return &apiv1.AddDeviceResponse{
		Device: protoND,
		Added:  true,
	}, nil
}

func (srv *server) DeleteDevice(ctx context.Context, req *apiv1.DeleteDeviceRequest) (*apiv1.DeleteDeviceResponse, error) {
	zlog.Info().Msgf("Removing network device (%s)", req.GetId())

	// sanity check for input parameters
	if req.GetId() == "" {
		err := fmt.Errorf("ID is not specified")
		zlog.Error().Err(err).Msg("Failed to delete network device")
		return nil, err
	}

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

func (srv *server) GetDeviceList(ctx context.Context, _ *emptypb.Empty) (*apiv1.GetDeviceListResponse, error) {
	zlog.Info().Msgf("Retrieving all available network devices")

	ndList, err := db.ListNetworkDevices(ctx, srv.dbClient)
	if err != nil {
		// failed to retrieve network devices
		return nil, err
	}
	// network devices were retrieved
	// converting list of network devices to proto notation
	protoNDlist := ConvertNetworkDeviceResourcesToNetworkDevicesProto(ndList)
	return &apiv1.GetDeviceListResponse{
		Devices: protoNDlist,
	}, nil
}

func (srv *server) UpdateDeviceList(ctx context.Context, req *apiv1.UpdateDeviceListRequest) (*apiv1.UpdateDeviceListResponse, error) {
	zlog.Info().Msgf("Updating network devices")

	retList := make([]*apiv1.NetworkDevice, 0)
	var cumulativeErr error
	for _, nd := range req.GetDevices() {
		protoND, err := srv.UpdateNetworkDevice(ctx, nd)
		if err != nil {
			zlog.Error().Err(err).Msg("Failed to update network device")
			cumulativeErr = errors.Join(cumulativeErr, err)
			continue
		}
		retList = append(retList, protoND)
	}
	if cumulativeErr != nil {
		// errors occurred during the update
		zlog.Info().Msgf("Errors occurred during bulk update of the network devices")
	}
	return &apiv1.UpdateDeviceListResponse{
		Devices: retList,
	}, cumulativeErr
}

func (srv *server) UpdateNetworkDevice(ctx context.Context, nd *apiv1.NetworkDevice) (*apiv1.NetworkDevice, error) {
	zlog.Info().Msgf("Updating network device (%s)", nd.GetId())

	entVendor := ConvertProtoVendorToEntVendor(nd.GetVendor())
	entEndpoints := ConvertProtoEndpointsToEndpoints(nd.GetEndpoints())
	updND, err := db.UpdateNetworkDeviceByUser(ctx, srv.dbClient, nd.GetId(), nd.GetModel(), entVendor, entEndpoints)
	if err != nil {
		return nil, err
	}

	protoND := ConvertNetworkDeviceResourceToNetworkDeviceProto(updND)
	return protoND, nil
}

func (srv *server) GetDeviceStatus(ctx context.Context, req *apiv1.GetDeviceStatusRequest) (*apiv1.GetDeviceStatusResponse, error) {
	zlog.Info().Msgf("Retrieving network device status (%s)", req.GetId())

	// first, retrieving network device resource by ID
	nd, err := db.GetNetworkDeviceByID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	// then, retrieving network device resource by its endpoint
	altNd, err := db.GetNetworkDeviceByEndpoint(ctx, srv.dbClient, req.GetEndpoint().GetHost(), req.GetEndpoint().GetPort())
	if err != nil {
		return nil, err
	}

	// now, comparing if it is the same device
	if !CompareNetworkDeviceResources(nd, altNd) {
		// resource violation is happening - network device resources are not identical
		newErr := fmt.Errorf("resource violation in the DB")
		zlog.Error().Err(newErr).Msgf("Network device resource violation in the DB: %v and %v", nd, altNd)
		return nil, newErr
	}

	// resources are identical, proceeding
	// retrieving device status
	s, err := db.GetDeviceStatusByNetworkDeviceID(ctx, srv.dbClient, req.GetId())
	if err != nil {
		return nil, err
	}

	protoStatus := ConvertEntDeviceStatusToProtoDeviceStatus(s)
	return &apiv1.GetDeviceStatusResponse{
		Id:       req.GetId(),
		Endpoint: req.GetEndpoint(),
		Status:   protoStatus,
	}, nil
}

func (srv *server) GetSummary(ctx context.Context, _ *emptypb.Empty) (*apiv1.GetSummaryResponse, error) {
	zlog.Info().Msgf("Retrieving network device summary")

	// retrieving all network devices currently available in the system
	ndList, err := srv.GetDeviceList(ctx, nil)
	if err != nil {
		return nil, err
	}

	resp := &apiv1.GetSummaryResponse{}
	var cumulativeErr error
	// fetching device status for each device and gathering statistics right away
	for _, nd := range ndList.GetDevices() {
		// fetching device status for at least one endpoint
		for _, ep := range nd.GetEndpoints() {
			ds, err := srv.GetDeviceStatus(ctx, CreateGetDeviceStatusRequest(nd.GetId(), ep))
			if err != nil {
				// aggregating errors
				cumulativeErr = errors.Join(cumulativeErr, err)
				continue
			}
			// device status was successfully fetched
			// gathering statistics
			resp = getStatistics(resp, ds.GetStatus())
			break // no need for further iterations
		}
	}
	if cumulativeErr != nil {
		zlog.Warn().Msgf("Errors occurred during network device summary collection")
	}
	return resp, cumulativeErr
}

func getStatistics(stats *apiv1.GetSummaryResponse, ds *apiv1.DeviceStatus) *apiv1.GetSummaryResponse {
	stats.DevicesTotal++
	switch ds.GetStatus() {
	case apiv1.Status_STATUS_DEVICE_UP:
		stats.DevicesUp++
	case apiv1.Status_STATUS_DEVICE_DOWN:
		stats.DownDevices++
	case apiv1.Status_STATUS_DEVICE_UNHEALTHY:
		stats.DevicesUnhealthy++
	}
	return stats
}

func (srv *server) SwapDeviceList(ctx context.Context, req *apiv1.SwapDeviceListRequest) (*apiv1.SwapDeviceListResponse, error) {
	zlog.Info().Msgf("Performing swap of the network devices in the controller")

	// performing initial sanity check
	if len(req.GetDevices()) == 0 {
		err := fmt.Errorf("at least one device is required")
		zlog.Error().Err(err).Msg("Swapping of network devices has failed")
		return nil, err
	}

	// retrieving list of existing devices
	existingNDList, err := srv.GetDeviceList(ctx, nil)
	if err != nil {
		return nil, err
	}

	// implementing dumb logic: deleting all existing devices and then creating new devices
	var cumulativeErr error
	for _, nd := range existingNDList.GetDevices() {
		resp, err := srv.DeleteDevice(ctx, CreateDeleteDeviceRequest(nd.GetId()))
		if err != nil {
			cumulativeErr = errors.Join(cumulativeErr, err)
		}
		if !resp.GetDeleted() {
			// device was not deleted, crafting error
			err = fmt.Errorf("network device (%s) was not deleted: %s", nd.GetId(), resp.GetDetails())
			cumulativeErr = errors.Join(cumulativeErr, err)
		}
	}
	if cumulativeErr != nil {
		zlog.Error().Err(cumulativeErr).Msgf("Errors occurred during network device deletion")
	}

	// now, creating new devices
	cumulativeErr = nil
	addedDevices := make([]*apiv1.NetworkDevice, 0)
	for _, nd := range req.GetDevices() {
		resp, err := srv.AddDevice(ctx, CreateAddDeviceRequest(nd.GetVendor(), nd.GetModel(), nd.GetEndpoints()))
		if err != nil {
			cumulativeErr = errors.Join(cumulativeErr, err)
		}
		if !resp.GetAdded() {
			err = fmt.Errorf("network device (%s) was not added: %s", nd.GetId(), resp.GetDetails())
			cumulativeErr = errors.Join(cumulativeErr, err)
		}
		addedDevices = append(addedDevices, resp.GetDevice())
	}
	if cumulativeErr != nil {
		zlog.Error().Err(cumulativeErr).Msgf("Errors occurred during network device addition")
	}
	if len(addedDevices) == 0 {
		err = fmt.Errorf("no network devices were added to the controller")
		zlog.Error().Err(err).Msgf("Device swap has failed")
		return nil, err
	}
	return &apiv1.SwapDeviceListResponse{
		Devices: addedDevices,
	}, nil
}

func (srv *server) GetAllDeviceStatuses(ctx context.Context, _ *emptypb.Empty) (*apiv1.GetAllDeviceStatusesResponse, error) {
	zlog.Info().Msgf("Retrieving all network device statuses")

	// listing all device statuses
	dss, err := db.ListDeviceStatuses(ctx, srv.dbClient)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to retrieve all network device statuses")
		return nil, err
	}

	retList := make([]*apiv1.DeviceStatus, 0)
	for _, ds := range dss {
		protoStatus := ConvertEntDeviceStatusToProtoDeviceStatus(ds)
		retList = append(retList, protoStatus)
	}
	return &apiv1.GetAllDeviceStatusesResponse{
		Statuses: retList,
	}, nil
}
