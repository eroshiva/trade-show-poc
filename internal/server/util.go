// Package server implements main server logic
package server

import (
	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/devicestatus"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
)

// ConvertNetworkDeviceResourcesToNetworkDevicesProto converts a list of ENT Network Device to Proto Network Device.
func ConvertNetworkDeviceResourcesToNetworkDevicesProto(nds []*ent.NetworkDevice) []*apiv1.NetworkDevice {
	ret := make([]*apiv1.NetworkDevice, 0)
	for _, nd := range nds {
		protoND := ConvertNetworkDeviceResourceToNetworkDeviceProto(nd)
		ret = append(ret, protoND)
	}
	return ret
}

// ConvertNetworkDeviceResourceToNetworkDeviceProto converts ENT Network Device to Proto Network Device.
func ConvertNetworkDeviceResourceToNetworkDeviceProto(nd *ent.NetworkDevice) *apiv1.NetworkDevice {
	sw := &apiv1.Version{}
	if nd.Edges.SwVersion != nil {
		sw = ConvertEntVersionToProtoVersion(nd.Edges.SwVersion)
	}
	fw := &apiv1.Version{}
	if nd.Edges.FwVersion != nil {
		fw = ConvertEntVersionToProtoVersion(nd.Edges.FwVersion)
	}

	ret := &apiv1.NetworkDevice{
		Id:        nd.ID,
		Vendor:    ConvertEntVendorToProtoVendor(nd.Vendor),
		Model:     nd.Model,
		Endpoints: make([]*apiv1.Endpoint, 0),
		HwVersion: nd.HwVersion,
		SwVersion: sw,
		FwVersion: fw,
	}

	endpoints := ConvertEndpointsToEndpointsProto(nd.Edges.Endpoints)
	ret.Endpoints = append(ret.Endpoints, endpoints...)
	return ret
}

// ConvertEntVendorToProtoVendor converts ENT Vendor to Proto Vendor.
func ConvertEntVendorToProtoVendor(vendor networkdevice.Vendor) apiv1.Vendor {
	switch vendor {
	case networkdevice.VendorVENDOR_UBIQUITI:
		return apiv1.Vendor_VENDOR_UBIQUITI
	case networkdevice.VendorVENDOR_CISCO:
		return apiv1.Vendor_VENDOR_CISCO
	case networkdevice.VendorVENDOR_JUNIPER:
		return apiv1.Vendor_VENDOR_JUNIPER
	default:
		return apiv1.Vendor_VENDOR_UNSPECIFIED
	}
}

// ConvertProtoVendorToEntVendor converts Proto vendor to ENT vendor.
func ConvertProtoVendorToEntVendor(vendor apiv1.Vendor) networkdevice.Vendor {
	switch vendor {
	case apiv1.Vendor_VENDOR_UBIQUITI:
		return networkdevice.VendorVENDOR_UBIQUITI
	case apiv1.Vendor_VENDOR_JUNIPER:
		return networkdevice.VendorVENDOR_JUNIPER
	case apiv1.Vendor_VENDOR_CISCO:
		return networkdevice.VendorVENDOR_CISCO
	default:
		return networkdevice.VendorVENDOR_UNSPECIFIED
	}
}

// ConvertEndpointsToEndpointsProto converts list of ENT endpoints to list of Proto endpoints.
func ConvertEndpointsToEndpointsProto(endpoints []*ent.Endpoint) []*apiv1.Endpoint {
	retList := make([]*apiv1.Endpoint, 0)
	if len(endpoints) > 0 {
		for _, ep := range endpoints {
			protoEndpoint := ConvertEndpointToEndpointProto(ep)
			retList = append(retList, protoEndpoint)
		}
	}
	return retList
}

// ConvertEndpointToEndpointProto converts ENT endpoint to Proto endpoint.
func ConvertEndpointToEndpointProto(endpoint *ent.Endpoint) *apiv1.Endpoint {
	return &apiv1.Endpoint{
		Id:       endpoint.ID,
		Host:     endpoint.Host,
		Port:     endpoint.Port,
		Protocol: ConvertEntProtocolToProtoProtocol(endpoint.Protocol),
	}
}

// ConvertProtoEndpointsToEndpoints converts list of Proto endpoint to list of ENT endpoint.
func ConvertProtoEndpointsToEndpoints(endpoints []*apiv1.Endpoint) []*ent.Endpoint {
	retList := make([]*ent.Endpoint, 0)
	if len(endpoints) > 0 {
		for _, ep := range endpoints {
			protoEndpoint := ConvertProtoEndpointToEndpoint(ep)
			retList = append(retList, protoEndpoint)
		}
	}
	return retList
}

// ConvertProtoEndpointToEndpoint converts Proto endpoint to ENT endpoint.
func ConvertProtoEndpointToEndpoint(endpoint *apiv1.Endpoint) *ent.Endpoint {
	return &ent.Endpoint{
		ID:       endpoint.GetId(),
		Host:     endpoint.GetHost(),
		Port:     endpoint.GetPort(),
		Protocol: ConvertProtoProtocolToEntProtocol(endpoint.GetProtocol()),
	}
}

// ConvertEntProtocolToProtoProtocol converts ENT Protocol to Proto Protocol.
func ConvertEntProtocolToProtoProtocol(protocol endpoint.Protocol) apiv1.Protocol {
	switch protocol {
	case endpoint.ProtocolPROTOCOL_NETCONF:
		return apiv1.Protocol_PROTOCOL_NETCONF
	case endpoint.ProtocolPROTOCOL_SNMP:
		return apiv1.Protocol_PROTOCOL_SNMP
	case endpoint.ProtocolPROTOCOL_RESTCONF:
		return apiv1.Protocol_PROTOCOL_RESTCONF
	case endpoint.ProtocolPROTOCOL_OPEN_V_SWITCH:
		return apiv1.Protocol_PROTOCOL_OPEN_V_SWITCH
	default:
		return apiv1.Protocol_PROTOCOL_UNSPECIFIED
	}
}

// ConvertProtoProtocolToEntProtocol converts Proto protocol to ENT protocol.
func ConvertProtoProtocolToEntProtocol(protocol apiv1.Protocol) endpoint.Protocol {
	switch protocol {
	case apiv1.Protocol_PROTOCOL_NETCONF:
		return endpoint.ProtocolPROTOCOL_NETCONF
	case apiv1.Protocol_PROTOCOL_SNMP:
		return endpoint.ProtocolPROTOCOL_SNMP
	case apiv1.Protocol_PROTOCOL_RESTCONF:
		return endpoint.ProtocolPROTOCOL_RESTCONF
	case apiv1.Protocol_PROTOCOL_OPEN_V_SWITCH:
		return endpoint.ProtocolPROTOCOL_OPEN_V_SWITCH
	default:
		return endpoint.ProtocolPROTOCOL_UNSPECIFIED
	}
}

// ConvertEntVersionToProtoVersion converts ENT version to Proto version.
func ConvertEntVersionToProtoVersion(version *ent.Version) *apiv1.Version {
	return &apiv1.Version{
		Id:       version.ID,
		Version:  version.Version,
		Checksum: version.Checksum,
	}
}

// CompareNetworkDeviceResources runs assertions on all fields of provided Network Device resources. Match is reported
// when Network Device resources are identical.
func CompareNetworkDeviceResources(nd1, nd2 *ent.NetworkDevice) bool {
	// running long ifs
	if nd1.ID != nd2.ID {
		return false
	}
	if nd1.Vendor != nd2.Vendor {
		return false
	}
	if nd1.Model != nd2.Model {
		return false
	}
	if nd1.HwVersion != nd2.HwVersion {
		return false
	}
	if nd1.Edges.SwVersion != nil && nd2.Edges.SwVersion != nil {
		if nd1.Edges.SwVersion.Version != nd2.Edges.SwVersion.Version {
			return false
		}
		if nd1.Edges.SwVersion.Checksum != nd2.Edges.SwVersion.Checksum {
			return false
		}
	}
	if nd1.Edges.FwVersion != nil && nd2.Edges.FwVersion != nil {
		if nd1.Edges.FwVersion.Version != nd2.Edges.FwVersion.Version {
			return false
		}
		if nd1.Edges.FwVersion.Checksum != nd2.Edges.FwVersion.Checksum {
			return false
		}
	}
	for _, nd1Endpoint := range nd1.Edges.Endpoints {
		found := false
		for _, nd2Endpoint := range nd2.Edges.Endpoints {
			if nd1Endpoint.ID == nd2Endpoint.ID {
				if nd1Endpoint.Protocol == nd2Endpoint.Protocol &&
					nd1Endpoint.Host == nd2Endpoint.Host &&
					nd1Endpoint.Port == nd2Endpoint.Port {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ConvertEntDeviceStatusToProtoDeviceStatus converts ENT Device Status to Proto Device Status notation.
func ConvertEntDeviceStatusToProtoDeviceStatus(ds *ent.DeviceStatus) *apiv1.DeviceStatus {
	return &apiv1.DeviceStatus{
		Id:       ds.ID,
		Status:   ConvertEntStatusToProtoStatus(ds.Status),
		LastSeen: ds.LastSeen,
	}
}

// ConvertEntStatusToProtoStatus converts ENT status to Proto status notation.
func ConvertEntStatusToProtoStatus(status devicestatus.Status) apiv1.Status {
	switch status {
	case devicestatus.StatusSTATUS_DEVICE_UP:
		return apiv1.Status_STATUS_DEVICE_UP
	case devicestatus.StatusSTATUS_DEVICE_UNHEALTHY:
		return apiv1.Status_STATUS_DEVICE_UNHEALTHY
	case devicestatus.StatusSTATUS_DEVICE_DOWN:
		return apiv1.Status_STATUS_DEVICE_DOWN
	default:
		return apiv1.Status_STATUS_UNSPECIFIED
	}
}
