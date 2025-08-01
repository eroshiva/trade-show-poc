// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/eroshiva/trade-show-poc/internal/ent/networkdevice"
	"github.com/eroshiva/trade-show-poc/internal/ent/version"
)

// NetworkDevice is the model entity for the NetworkDevice schema.
type NetworkDevice struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// Vendor holds the value of the "vendor" field.
	Vendor networkdevice.Vendor `json:"vendor,omitempty"`
	// Model holds the value of the "model" field.
	Model string `json:"model,omitempty"`
	// HwVersion holds the value of the "hw_version" field.
	HwVersion string `json:"hw_version,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the NetworkDeviceQuery when eager-loading is set.
	Edges                     NetworkDeviceEdges `json:"edges"`
	network_device_sw_version *string
	network_device_fw_version *string
	selectValues              sql.SelectValues
}

// NetworkDeviceEdges holds the relations/edges for other nodes in the graph.
type NetworkDeviceEdges struct {
	// Endpoints holds the value of the endpoints edge.
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
	// SwVersion holds the value of the sw_version edge.
	SwVersion *Version `json:"sw_version,omitempty"`
	// FwVersion holds the value of the fw_version edge.
	FwVersion *Version `json:"fw_version,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// EndpointsOrErr returns the Endpoints value or an error if the edge
// was not loaded in eager-loading.
func (e NetworkDeviceEdges) EndpointsOrErr() ([]*Endpoint, error) {
	if e.loadedTypes[0] {
		return e.Endpoints, nil
	}
	return nil, &NotLoadedError{edge: "endpoints"}
}

// SwVersionOrErr returns the SwVersion value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e NetworkDeviceEdges) SwVersionOrErr() (*Version, error) {
	if e.SwVersion != nil {
		return e.SwVersion, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: version.Label}
	}
	return nil, &NotLoadedError{edge: "sw_version"}
}

// FwVersionOrErr returns the FwVersion value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e NetworkDeviceEdges) FwVersionOrErr() (*Version, error) {
	if e.FwVersion != nil {
		return e.FwVersion, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: version.Label}
	}
	return nil, &NotLoadedError{edge: "fw_version"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*NetworkDevice) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case networkdevice.FieldID, networkdevice.FieldVendor, networkdevice.FieldModel, networkdevice.FieldHwVersion:
			values[i] = new(sql.NullString)
		case networkdevice.ForeignKeys[0]: // network_device_sw_version
			values[i] = new(sql.NullString)
		case networkdevice.ForeignKeys[1]: // network_device_fw_version
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the NetworkDevice fields.
func (nd *NetworkDevice) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case networkdevice.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				nd.ID = value.String
			}
		case networkdevice.FieldVendor:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field vendor", values[i])
			} else if value.Valid {
				nd.Vendor = networkdevice.Vendor(value.String)
			}
		case networkdevice.FieldModel:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field model", values[i])
			} else if value.Valid {
				nd.Model = value.String
			}
		case networkdevice.FieldHwVersion:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field hw_version", values[i])
			} else if value.Valid {
				nd.HwVersion = value.String
			}
		case networkdevice.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field network_device_sw_version", values[i])
			} else if value.Valid {
				nd.network_device_sw_version = new(string)
				*nd.network_device_sw_version = value.String
			}
		case networkdevice.ForeignKeys[1]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field network_device_fw_version", values[i])
			} else if value.Valid {
				nd.network_device_fw_version = new(string)
				*nd.network_device_fw_version = value.String
			}
		default:
			nd.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the NetworkDevice.
// This includes values selected through modifiers, order, etc.
func (nd *NetworkDevice) Value(name string) (ent.Value, error) {
	return nd.selectValues.Get(name)
}

// QueryEndpoints queries the "endpoints" edge of the NetworkDevice entity.
func (nd *NetworkDevice) QueryEndpoints() *EndpointQuery {
	return NewNetworkDeviceClient(nd.config).QueryEndpoints(nd)
}

// QuerySwVersion queries the "sw_version" edge of the NetworkDevice entity.
func (nd *NetworkDevice) QuerySwVersion() *VersionQuery {
	return NewNetworkDeviceClient(nd.config).QuerySwVersion(nd)
}

// QueryFwVersion queries the "fw_version" edge of the NetworkDevice entity.
func (nd *NetworkDevice) QueryFwVersion() *VersionQuery {
	return NewNetworkDeviceClient(nd.config).QueryFwVersion(nd)
}

// Update returns a builder for updating this NetworkDevice.
// Note that you need to call NetworkDevice.Unwrap() before calling this method if this NetworkDevice
// was returned from a transaction, and the transaction was committed or rolled back.
func (nd *NetworkDevice) Update() *NetworkDeviceUpdateOne {
	return NewNetworkDeviceClient(nd.config).UpdateOne(nd)
}

// Unwrap unwraps the NetworkDevice entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (nd *NetworkDevice) Unwrap() *NetworkDevice {
	_tx, ok := nd.config.driver.(*txDriver)
	if !ok {
		panic("ent: NetworkDevice is not a transactional entity")
	}
	nd.config.driver = _tx.drv
	return nd
}

// String implements the fmt.Stringer.
func (nd *NetworkDevice) String() string {
	var builder strings.Builder
	builder.WriteString("NetworkDevice(")
	builder.WriteString(fmt.Sprintf("id=%v, ", nd.ID))
	builder.WriteString("vendor=")
	builder.WriteString(fmt.Sprintf("%v", nd.Vendor))
	builder.WriteString(", ")
	builder.WriteString("model=")
	builder.WriteString(nd.Model)
	builder.WriteString(", ")
	builder.WriteString("hw_version=")
	builder.WriteString(nd.HwVersion)
	builder.WriteByte(')')
	return builder.String()
}

// NetworkDevices is a parsable slice of NetworkDevice.
type NetworkDevices []*NetworkDevice
