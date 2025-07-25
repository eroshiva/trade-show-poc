// File updated by protoc-gen-ent.

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Version struct {
	ent.Schema
}

func (Version) Fields() []ent.Field {
	return []ent.Field{field.String("version"), field.String("checksum")}
}
func (Version) Edges() []ent.Edge {
	return nil
}
func (Version) Annotations() []schema.Annotation {
	return nil
}
