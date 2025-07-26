// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type DeviceStatus struct {
	ent.Schema
}

func (DeviceStatus) Fields() []ent.Field {
	return []ent.Field{field.String("id"), field.Enum("status").Values("STATUS_UNSPECIFIED", "STATUS_DEVICE_DOWN", "STATUS_DEVICE_UNHEALTHY", "STATUS_DEVICE_UP"), field.String("last_seen")}
}
func (DeviceStatus) Edges() []ent.Edge {
	return []ent.Edge{edge.To("network_device", NetworkDevice.Type).Unique()}
}
func (DeviceStatus) Annotations() []schema.Annotation {
	return nil
}
