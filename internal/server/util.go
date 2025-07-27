// Package server implements main server logic
package server

import (
	apiv1 "github.com/eroshiva/trade-show-poc/api/v1"
	"github.com/eroshiva/trade-show-poc/internal/ent"
	"github.com/eroshiva/trade-show-poc/internal/ent/endpoint"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
)

// ConvertNetworkDeviceResourceToNetworkDeviceProto converts ENT Network Device to Proto Network Device.
func ConvertNetworkDeviceResourceToNetworkDeviceProto(nd *ent.NetworkDevice) *apiv1.NetworkDevice {
	endpoints := ConvertEndpointsToEndpointsProto(nd.Edges.Endpoints)
	sw := ConvertEntVersionToProtoVersion(nd.Edges.SwVersion)
	fw := ConvertEntVersionToProtoVersion(nd.Edges.FwVersion)
	return &apiv1.NetworkDevice{
		Id:        nd.ID,
		Vendor:    ConvertEntVendorToProtoVendor(nd.Vendor),
		Model:     nd.Model,
		Endpoints: endpoints,
		HwVersion: nd.HwVersion,
		SwVersion: sw,
		FwVersion: fw,
	}
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
