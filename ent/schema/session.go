// ent/schema/session.go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"go.jetify.com/typeid/v2"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Mixin of the Session.
func (Session) Mixin() []ent.Mixin {
	return []ent.Mixin{
		// created_at and updated_at (for last_active)
		mixin.Time{},
	}
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(func() string {
				return typeid.MustGenerate("sess").String()
			}).
			NotEmpty().
			Unique().
			// Sensitive().
			Immutable(),
		field.Time("expires_at").
			Default(time.Now().Add(24 * time.Hour)), // Default expiration to a day, if remember me, then 30 days
		field.String("ip_address").
			Optional().
			Nillable(),
		field.String("user_agent").
			Optional().
			Nillable(),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type). // Each user can have multiple sessions.
						Ref("sessions").
						Unique().    // Each session belongs to exactly one user.
						Required().  // The user ID is required when creating a session.
						Immutable(), // The user ID cannot be changed after the session is created.
	}
}

// Indexes of the Session.
func (Session) Indexes() []ent.Index {
	return []ent.Index{
		// Index for efficiently querying and cleaning up expired sessions.
		index.Fields("expires_at"),
	}
}
