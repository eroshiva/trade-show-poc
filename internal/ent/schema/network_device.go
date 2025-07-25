// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type NetworkDevice struct {
	ent.Schema
}

func (NetworkDevice) Fields() []ent.Field {
	return []ent.Field{field.String("id"), field.Enum("vendor").Values("VENDOR_UNSPECIFIED", "VENDOR_UBIQUITI", "VENDOR_CISCO", "VENDOR_JUNIPER"), field.String("model"), field.String("hw_version")}
}
func (NetworkDevice) Edges() []ent.Edge {
	return []ent.Edge{edge.To("endpoints", Endpoint.Type), edge.To("sw_version", Version.Type), edge.To("fw_version", Version.Type)}
}
func (NetworkDevice) Annotations() []schema.Annotation {
	return nil
}
