package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ChatMember holds the schema definition for the ChatMember entity.
type ChatMember struct {
	ent.Schema
}

// Fields of the ChatMember.
func (ChatMember) Fields() []ent.Field {
	return []ent.Field{
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
		field.Bool("is_admin").
			Default(false),
	}
}

// Edges of the ChatMember.
func (ChatMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("chat_members").
			Unique().
			Required(),
		edge.From("chat", Chat.Type).
			Ref("members").
			Unique().
			Required(),
	}
}
