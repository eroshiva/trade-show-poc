// Package server implements initial means for client-server communication
package server

import apiv1 "github.com/eroshiva/trade-show-poc/api/v1"

// CreateAddDeviceRequest is a helper wrapper function that creates an AddDeviceRequest message.
func CreateAddDeviceRequest(vendor apiv1.Vendor, model string, endpoints []*apiv1.Endpoint) *apiv1.AddDeviceRequest {
	return &apiv1.AddDeviceRequest{
		Device: &apiv1.NetworkDevice{
			Vendor:    vendor,
			Model:     model,
			Endpoints: endpoints,
		},
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
