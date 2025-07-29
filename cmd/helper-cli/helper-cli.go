// Package main is a main entry point for the helper gRPC CLI utility.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/server"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	component     = "component"
	componentName = "helper-cli"

	defaultTimeout = 1 * time.Second
	configFileName = "./cmd/helper-cli/config.json"
	defaultEPHost  = "localhost" // in case of k8s deployment, should be something like device-simulator-0.device-simulator-svc.monitoring-system.svc.cluster.local
	defaultEPPort  = "50151"

	snmp     = "SNMP"
	netconf  = "NETCONF"
	restconf = "RESTCONF"
	ovs      = "OVS"

	ubiquiti = "UBIQUITI"
	juniper  = "JUNIPER"
	cisco    = "CISCO"
)

var (
	zlog = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			return filepath.Dir(fmt.Sprintf("%s/", i))
		},
	}).Level(zerolog.TraceLevel).With().Caller().Timestamp().Str(component, componentName).Logger()

	grpcServerAddress = flag.String(
		"grpcServerAddress",
		"localhost:50051",
		"gRPC server address of the monitoring service to connect to",
	)

	// specifying various flags
	addDeviceFlag  = "addDevice"
	addDevice      = flag.Bool(addDeviceFlag, false, "Adds a default network device to the controller")
	addDevicesFlag = "addDevices"
	addDevices     = flag.Bool(addDevicesFlag, false, "Adds network devices (as specified in the JSON config file)")

	// deleting device
	deleteDeviceFlag   = "deleteDevice"
	deleteDevice       = flag.Bool(deleteDeviceFlag, false, "Deletes a network device from the controller")
	deleteDeviceIDFlag = "deleteDeviceID"
	deleteDeviceID     = flag.String(deleteDeviceIDFlag, "", "Specifies network device ID that needs to be deleted")

	// deleting all devices
	deleteAllDevicesFlag = "deleteAllDevices"
	deleteAllDevices     = flag.Bool(deleteAllDevicesFlag, false, "Deletes all network devices from the controller")

	// getting status of a specific device
	getStatusFlag      = "getStatus"
	getStatus          = flag.Bool(getStatusFlag, false, "Gets the status of the device")
	deviceIDFlag       = "deviceID"
	deviceID           = flag.String(deviceIDFlag, "", "Specifies the device ID")
	getAllStatusesFlag = "getAllStatuses"
	getAllStatuses     = flag.Bool(getAllStatusesFlag, false, "Gets all statuses of all of the network devices present in the system")
	getSummaryFlag     = "getSummary"
	getSummary         = flag.Bool(getSummaryFlag, false, "Gets the summary of the network devices present in the system")

	// updating list of the devices
	updateDevicesFlag = "updateDevices"
	updateDevices     = flag.Bool(updateDevicesFlag, false, "Updates all network devices in the system (as specified in the JSON config file)")
	swapDevicesFlag   = "swapDevices"
	swapDevices       = flag.Bool(swapDevicesFlag, false, "Swaps devices present in the system with a set of new ones (as specified in the JSON config file). "+
		"All network devices that are not present in this list are deleted from the monitoring system.")
)

func main() {
	zlog.Info().Msg("Starting helper-cli utility")
	flag.Parse()

	// creating the gRPC client
	conn, err := grpc.NewClient(
		*grpcServerAddress, // The address of the gRPC server
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zlog.Fatal().Err(err).Msgf("Failed to dial to gRPC server %s", *grpcServerAddress)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to close gRPC connection")
		}
	}()
	grpcClient := apiv1.NewDeviceMonitoringServiceClient(conn)

	if *addDevice {
		err = addNetworkDevice(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to add network device to the controller")
		}
	}

	if *addDevices {
		err = addNetworkDevices(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to add network devices to the controller")
		}
	}

	if *deleteDevice {
		err = deleteNetworkDevice(grpcClient, *deleteDeviceID)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to delete network device from the controller")
		}
	}

	if *deleteAllDevices {
		err = deleteAllNetworkDevices(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to delete all network devices from the controller")
		}
	}

	if *getStatus {
		err = getDeviceStatus(grpcClient, *deviceID)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to get device status")
		}
	}

	if *getAllStatuses {
		err = getAllNDStatuses(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to get all network device statuses")
		}
	}

	if *updateDevices {
		err = updateNetworkDevices(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to update network devices in the controller")
		}
	}

	if *swapDevices {
		err = swapNetworkDevices(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to swap network devices in the controller")
		}
	}

	if *getSummary {
		err = retrieveSummary(grpcClient)
		if err != nil {
			zlog.Error().Err(err).Msgf("Failed to retrieve device summary")
		}
	}
}

// addNetworkDevice adds default network device to the monitoring system.
func addNetworkDevice(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	ep := server.CreateEndpoint(defaultEPHost, defaultEPPort, apiv1.Protocol_PROTOCOL_NETCONF)
	req := server.CreateAddDeviceRequest(apiv1.Vendor_VENDOR_UBIQUITI, "XYZ", []*apiv1.Endpoint{ep})

	nd, err := grpcClient.AddDevice(ctx, req)
	if err != nil {
		return err
	}
	zlog.Info().Msgf("Successfully added network device to the controller: %v", nd)
	return nil
}

// Endpoint defines a type for parsing endpoints
type Endpoint struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
}

type NetworkDevice struct {
	Vendor   string      `json:"vendor"`
	Model    string      `json:"model"`
	Endpoint []*Endpoint `json:"endpoints"`
}

// addNetworkDevices reads a list of network devices specified in JSON config file.
func addNetworkDevices(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	// unmarshalling JSON file to Golang struct
	devices, err := parseJSON()
	if err != nil {
		return err
	}
	nds := convertGolangStructsToProtoStructs(devices)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var cumulativeErr error
	for _, nd := range nds {
		// run an add network device call
		resp, err := grpcClient.AddDevice(ctx, server.CreateAddDeviceRequest(nd.GetVendor(), nd.GetModel(), nd.GetEndpoints()))
		if err != nil {
			cumulativeErr = errors.Join(cumulativeErr, err)
			continue
		}
		if !resp.GetAdded() {
			err := fmt.Errorf("failed to add network device (%v) to the controller: %v", nd, resp.GetDetails())
			cumulativeErr = errors.Join(cumulativeErr, err)
		}
	}

	return cumulativeErr
}

func parseJSON() ([]*NetworkDevice, error) {
	configFile, err := os.Open(configFileName)
	if err != nil {
		zlog.Error().Err(err).Msgf("Failed to open config file")
		return nil, err
	}

	var devices []*NetworkDevice
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&devices); err != nil {
		zlog.Error().Err(err).Msgf("Failed to decode config file")
		return nil, err
	}
	return devices, nil
}

func convertGolangStructsToProtoStructs(devices []*NetworkDevice) []*apiv1.NetworkDevice {
	ret := make([]*apiv1.NetworkDevice, 0)
	for _, nd := range devices {
		eps := make([]*apiv1.Endpoint, 0)
		for _, ep := range nd.Endpoint {
			convProto := convertProtocol(ep.Protocol)
			convEp := server.CreateEndpoint(ep.Host, ep.Port, convProto)
			eps = append(eps, convEp)
		}
		convND := server.CreateNetworkDevice(convertVendor(nd.Vendor), nd.Model, eps)
		ret = append(ret, convND)
	}
	return ret
}

func convertProtocol(protocol string) apiv1.Protocol {
	switch strings.ToLower(protocol) {
	case strings.ToLower(snmp):
		return apiv1.Protocol_PROTOCOL_SNMP
	case strings.ToLower(netconf):
		return apiv1.Protocol_PROTOCOL_NETCONF
	case strings.ToLower(restconf):
		return apiv1.Protocol_PROTOCOL_RESTCONF
	case strings.ToLower(ovs):
		return apiv1.Protocol_PROTOCOL_OPEN_V_SWITCH
	default:
		return apiv1.Protocol_PROTOCOL_UNSPECIFIED
	}
}

func convertVendor(vendor string) apiv1.Vendor {
	switch strings.ToLower(vendor) {
	case strings.ToLower(ubiquiti):
		return apiv1.Vendor_VENDOR_UBIQUITI
	case strings.ToLower(juniper):
		return apiv1.Vendor_VENDOR_JUNIPER
	case strings.ToLower(cisco):
		return apiv1.Vendor_VENDOR_CISCO
	default:
		return apiv1.Vendor_VENDOR_UNSPECIFIED
	}
}

func deleteNetworkDevice(grpcClient apiv1.DeviceMonitoringServiceClient, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := grpcClient.DeleteDevice(ctx, server.CreateDeleteDeviceRequest(id))
	return err
}

func deleteAllNetworkDevices(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// retrieving list of network devices
	ndList, err := grpcClient.GetDeviceList(ctx, nil)
	if err != nil {
		return err
	}

	var cumulativeErr error
	for _, nd := range ndList.GetDevices() {
		err = deleteNetworkDevice(grpcClient, nd.GetId())
		if err != nil {
			cumulativeErr = errors.Join(cumulativeErr, err)
		}
	}

	return cumulativeErr
}

func getDeviceStatus(grpcClient apiv1.DeviceMonitoringServiceClient, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	ndList, err := grpcClient.GetDeviceList(ctx, nil)
	if err != nil {
		return err
	}

	eps := make([]*apiv1.Endpoint, 0)
	for _, nd := range ndList.GetDevices() {
		if nd.GetId() == id {
			// device match is found, saving endpoint
			eps = nd.GetEndpoints()
		}
	}
	if len(eps) == 0 {
		err = fmt.Errorf("failed to find endpoints for the network device (%s)", id)
		return err
	}

	// taking only first endpoint
	ds, err := grpcClient.GetDeviceStatus(ctx, server.CreateGetDeviceStatusRequest(id, eps[0]))
	if err != nil {
		return err
	}
	zlog.Info().Msgf("Retrieved device status for network device (%s): %s at %s", id, ds.GetStatus().GetStatus(), ds.GetStatus().GetLastSeen())
	return nil
}

func getAllNDStatuses(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	ndList, err := grpcClient.GetDeviceList(ctx, nil)
	if err != nil {
		return err
	}

	var cumulativeErr error
	for _, nd := range ndList.GetDevices() {
		if len(nd.GetEndpoints()) == 0 {
			err = fmt.Errorf("failed to find endpoints for the network device (%s)", nd.GetId())
			cumulativeErr = errors.Join(cumulativeErr, err)
			continue
		}
		ds, err := grpcClient.GetDeviceStatus(ctx, server.CreateGetDeviceStatusRequest(nd.GetId(), nd.GetEndpoints()[0]))
		if err != nil {
			cumulativeErr = errors.Join(cumulativeErr, err)
			continue
		}
		zlog.Info().Msgf("Retrieved device status for network device (%s): %s at %s", nd.GetId(), ds.GetStatus().GetStatus(), ds.GetStatus().GetLastSeen())
	}
	return cumulativeErr
}

func updateNetworkDevices(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	// unmarshalling JSON file to Golang struct
	devices, err := parseJSON()
	if err != nil {
		return err
	}
	nds := convertGolangStructsToProtoStructs(devices)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = grpcClient.UpdateDeviceList(ctx, server.CreateUpdateDeviceListRequest(nds))
	if err != nil {
		return err
	}
	zlog.Info().Msg("Network devices were updated")
	return nil
}

func swapNetworkDevices(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	// unmarshalling JSON file to Golang struct
	devices, err := parseJSON()
	if err != nil {
		return err
	}
	nds := convertGolangStructsToProtoStructs(devices)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = grpcClient.SwapDeviceList(ctx, server.CreateSwapDeviceListRequest(nds))
	if err != nil {
		return err
	}
	zlog.Info().Msg("Network devices were swapped")
	return nil
}

func retrieveSummary(grpcClient apiv1.DeviceMonitoringServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	summary, err := grpcClient.GetSummary(ctx, nil)
	if err != nil {
		return err
	}
	zlog.Info().Msg("Retrieved summary for all network devices")
	zlog.Info().Msgf("Total number of devices in the system: %d", summary.GetDevicesTotal())
	zlog.Info().Msgf("Number of devices in UP state: %d", summary.GetDevicesUp())
	zlog.Info().Msgf("Number of devices in UNHEALTHY state: %d", summary.GetDevicesUnhealthy())
	zlog.Info().Msgf("Number of devices in DOWN state: %d", summary.GetDownDevices())
	return nil
}
