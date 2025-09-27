package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"go.jetify.com/typeid/v2"
)

// Like holds the schema definition for the Like entity.
type Like struct {
	ent.Schema
}

// Mixin of the Like.
func (Like) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{}, // Provides created_at and updated_at fields
	}
}

// Fields of the Like.
func (Like) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return typeid.MustGenerate("like").String()
			}).
			NotEmpty().
			Unique().
			Immutable(),
		field.String("user_id").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

// Edges of the Like.
func (Like) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("project", Project.Type).
			Unique().
			Required().
			Field("project_id"),
	}
}

// Indexes of the Like.
func (Like) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "project_id").
			Unique(),
		index.Fields("project_id"),
	}
}