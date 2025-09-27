package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"go.jetify.com/typeid/v2"
)

// ProjectTag holds the schema definition for the ProjectTag entity.
type ProjectTag struct {
	ent.Schema
}

// Mixin of the ProjectTag.
func (ProjectTag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{}, // Provides created_at and updated_at fields
	}
}

// Fields of the ProjectTag.
func (ProjectTag) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return typeid.MustGenerate("ptag").String()
			}).
			NotEmpty().
			Unique().
			Immutable(),
		field.String("project_id").
			NotEmpty(),
		field.String("tag_id").
			NotEmpty(),
	}
}

// Edges of the ProjectTag.
func (ProjectTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("project", Project.Type).
			Unique().
			Required().
			Field("project_id"),
		edge.To("tag", Tag.Type).
			Unique().
			Required().
			Field("tag_id"),
	}
}

// Indexes of the ProjectTag.
func (ProjectTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "tag_id").
			Unique(),
		index.Fields("tag_id"),
	}
}