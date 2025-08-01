// Code generated by ent, DO NOT EDIT.

package devicestatus

import (
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the devicestatus type in the database.
	Label = "device_status"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldLastSeen holds the string denoting the last_seen field in the database.
	FieldLastSeen = "last_seen"
	// FieldConsequentialFailedConnectivityAttempts holds the string denoting the consequential_failed_connectivity_attempts field in the database.
	FieldConsequentialFailedConnectivityAttempts = "consequential_failed_connectivity_attempts"
	// EdgeNetworkDevice holds the string denoting the network_device edge name in mutations.
	EdgeNetworkDevice = "network_device"
	// Table holds the table name of the devicestatus in the database.
	Table = "device_status"
	// NetworkDeviceTable is the table that holds the network_device relation/edge.
	NetworkDeviceTable = "device_status"
	// NetworkDeviceInverseTable is the table name for the NetworkDevice entity.
	// It exists in this package in order to avoid circular dependency with the "networkdevice" package.
	NetworkDeviceInverseTable = "network_devices"
	// NetworkDeviceColumn is the table column denoting the network_device relation/edge.
	NetworkDeviceColumn = "device_status_network_device"
)

// Columns holds all SQL columns for devicestatus fields.
var Columns = []string{
	FieldID,
	FieldStatus,
	FieldLastSeen,
	FieldConsequentialFailedConnectivityAttempts,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "device_status"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"device_status_network_device",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

// Status defines the type for the "status" enum field.
type Status string

// Status values.
const (
	StatusSTATUS_UNSPECIFIED      Status = "STATUS_UNSPECIFIED"
	StatusSTATUS_DEVICE_DOWN      Status = "STATUS_DEVICE_DOWN"
	StatusSTATUS_DEVICE_UNHEALTHY Status = "STATUS_DEVICE_UNHEALTHY"
	StatusSTATUS_DEVICE_UP        Status = "STATUS_DEVICE_UP"
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusSTATUS_UNSPECIFIED, StatusSTATUS_DEVICE_DOWN, StatusSTATUS_DEVICE_UNHEALTHY, StatusSTATUS_DEVICE_UP:
		return nil
	default:
		return fmt.Errorf("devicestatus: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the DeviceStatus queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByLastSeen orders the results by the last_seen field.
func ByLastSeen(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLastSeen, opts...).ToFunc()
}

// ByConsequentialFailedConnectivityAttempts orders the results by the consequential_failed_connectivity_attempts field.
func ByConsequentialFailedConnectivityAttempts(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldConsequentialFailedConnectivityAttempts, opts...).ToFunc()
}

// ByNetworkDeviceField orders the results by network_device field.
func ByNetworkDeviceField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newNetworkDeviceStep(), sql.OrderByField(field, opts...))
	}
}
func newNetworkDeviceStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(NetworkDeviceInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, NetworkDeviceTable, NetworkDeviceColumn),
	)
}
