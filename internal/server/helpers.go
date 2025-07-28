// Package server implements initial means for client-server communication
package server

import (
	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
)

// CreateAddDeviceRequest is a helper wrapper function that creates an AddDeviceRequest message.
func CreateAddDeviceRequest(vendor apiv1.Vendor, model string, endpoints []*apiv1.Endpoint) *apiv1.AddDeviceRequest {
	return &apiv1.AddDeviceRequest{
		Device: CreateNetworkDevice(vendor, model, endpoints),
	}
}

// CreateEndpoint is a helper wrapper function that creates Endpoint message.
func CreateEndpoint(host, port string, protocol apiv1.Protocol) *apiv1.Endpoint {
	return &apiv1.Endpoint{
		Host:     host,
		Port:     port,
		Protocol: protocol,
	}
}

// CreateVersion is a helper wrapper function that creates Version message.
func CreateVersion(version, checksum string) *apiv1.Version {
	return &apiv1.Version{
		Version:  version,
		Checksum: checksum,
	}
}

// DeviceStatus is a helper wrapper function that creates DeviceStatus message.
func DeviceStatus(status apiv1.Status, lastSeen string) *apiv1.DeviceStatus {
	return &apiv1.DeviceStatus{
		Status:   status,
		LastSeen: lastSeen,
	}
}

// CreateDeleteDeviceRequest is a helper wrapper function that creates DeleteDeviceRequest message.
func CreateDeleteDeviceRequest(id string) *apiv1.DeleteDeviceRequest {
	return &apiv1.DeleteDeviceRequest{
		Id: id,
	}
}

// CreateNetworkDevice is a helper wrapper function that creates NetworkDevice.
func CreateNetworkDevice(vendor apiv1.Vendor, model string, endpoints []*apiv1.Endpoint) *apiv1.NetworkDevice {
	return &apiv1.NetworkDevice{
		Vendor:    vendor,
		Model:     model,
		Endpoints: endpoints,
	}
}

// CreateUpdateDeviceListRequest is a helper wrapper function that creates UpdateDeviceListRequest message.
func CreateUpdateDeviceListRequest(nds []*apiv1.NetworkDevice) *apiv1.UpdateDeviceListRequest {
	return &apiv1.UpdateDeviceListRequest{
		Devices: nds,
	}
}

// CreateSwapDeviceListRequest is a helper wrapper function that creates SwapDeviceListRequest message.
func CreateSwapDeviceListRequest(nds []*apiv1.NetworkDevice) *apiv1.SwapDeviceListRequest {
	return &apiv1.SwapDeviceListRequest{
		Devices: nds,
	}
}

// CreateGetDeviceStatusRequest is a helper wrapper that creates a GetDeviceStatusRequest message.
func CreateGetDeviceStatusRequest(id string, ep *apiv1.Endpoint) *apiv1.GetDeviceStatusRequest {
	return &apiv1.GetDeviceStatusRequest{
		Id:       id,
		Endpoint: ep,
	}
}
