package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"go.jetify.com/typeid/v2"
)

// UserTechnology holds the schema definition for the UserTechnology entity.
type UserTechnology struct {
	ent.Schema
}

// Mixin of the UserTechnology.
func (UserTechnology) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{}, // Provides created_at and updated_at fields
	}
}

// Fields of the UserTechnology.
func (UserTechnology) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return typeid.MustGenerate("utech").String()
			}).
			NotEmpty().
			Unique().
			Immutable(),
		field.String("user_id").
			NotEmpty(),
		field.String("technology_id").
			NotEmpty(),
		field.Enum("skill_level").
			Values("beginner", "intermediate", "expert").
			Default("beginner"),
		field.Float("years_experience").
			Optional().
			Nillable(),
	}
}

// Edges of the UserTechnology.
func (UserTechnology) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Unique().
			Required().
			Field("user_id"),
		edge.To("technology", Tag.Type).
			Unique().
			Required().
			Field("technology_id"),
	}
}

// Indexes of the UserTechnology.
func (UserTechnology) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "technology_id").
			Unique(),
		index.Fields("technology_id"),
		index.Fields("skill_level"),
	}
}
