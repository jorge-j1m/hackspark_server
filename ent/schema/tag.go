package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"go.jetify.com/typeid/v2"
)

// Tag holds the schema definition for the Tag entity.
type Tag struct {
	ent.Schema
}

// Mixin of the Tag.
func (Tag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{}, // Provides created_at and updated_at fields
	}
}

// Fields of the Tag.
func (Tag) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return typeid.MustGenerate("tag").String()
			}).
			NotEmpty().
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("slug").
			NotEmpty().
			Unique(),
		field.String("icon").
			Optional().
			Nillable(),
		field.String("description").
			Optional().
			Nillable(),
		field.Enum("category").
			Values("language", "framework", "tool", "database", "other").
			Default("other"),
		field.Int("usage_count").
			Default(0),
	}
}

// Edges of the Tag.
func (Tag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("creator", User.Type).
			Ref("created_tags").
			Unique(),
		edge.From("projects", Project.Type).
			Ref("tags").
			Through("project_tags", ProjectTag.Type),
		edge.From("users", User.Type).
			Ref("technologies").
			Through("user_technologies", UserTechnology.Type),
	}
}

// Indexes of the Tag.
func (Tag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").
			Unique(),
		index.Fields("usage_count"),
		index.Fields("name"),
		index.Fields("category"),
	}
}