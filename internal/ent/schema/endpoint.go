// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Endpoint struct {
	ent.Schema
}

func (Endpoint) Fields() []ent.Field {
	return []ent.Field{field.String("host"), field.String("port"), field.Enum("protocol").Values("PROTOCOL_UNSPECIFIED", "PROTOCOL_SNMP", "PROTOCOL_NETCONF", "PROTOCOL_RESTCONF", "PROTOCOL_OPEN_V_SWITCH")}
}
func (Endpoint) Edges() []ent.Edge {
	return []ent.Edge{edge.From("network_device", NetworkDevice.Type).Ref("endpoint")}
}
func (Endpoint) Annotations() []schema.Annotation {
	return nil
}
