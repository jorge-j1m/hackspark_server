package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"go.jetify.com/typeid/v2"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Mixin of the Project.
func (Project) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{}, // Provides created_at and updated_at fields
	}
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return typeid.MustGenerate("proj").String()
			}).
			NotEmpty().
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.Int("like_count").
			Default(0),
		field.Int("star_count").
			Default(0),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("owned_projects").
			Unique().
			Required(),
		edge.From("liked_by", User.Type).
			Ref("liked_projects").
			Through("likes", Like.Type),
		edge.To("tags", Tag.Type).
			Through("project_tags", ProjectTag.Type),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("like_count"),
		index.Edges("owner"),
	}
}
