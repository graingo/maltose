package mtrace

import (
	"context"

	"github.com/graingo/maltose/container/mvar"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/baggage"
)

// Baggage is a mechanism for propagating key-value data in a distributed system.
// It allows attaching custom data (such as user ID, request ID, etc.) to traces and propagating them across service calls.
type Baggage struct {
	ctx context.Context
}

// NewBaggage creates a new Baggage instance.
func NewBaggage(ctx context.Context) *Baggage {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Baggage{
		ctx: ctx,
	}
}

// SetValue sets a single baggage value.
func (b *Baggage) SetValue(key string, value interface{}) context.Context {
	member, _ := baggage.NewMember(key, cast.ToString(value))
	bag, _ := baggage.New(member)
	b.ctx = baggage.ContextWithBaggage(b.ctx, bag)
	return b.ctx
}

// SetMap sets multiple baggage values.
func (b *Baggage) SetMap(data map[string]interface{}) context.Context {
	members := make([]baggage.Member, 0)
	for k, v := range data {
		member, _ := baggage.NewMember(k, cast.ToString(v))
		members = append(members, member)
	}
	bag, _ := baggage.New(members...)
	b.ctx = baggage.ContextWithBaggage(b.ctx, bag)
	return b.ctx
}

// GetMap gets all baggage values.
func (b *Baggage) GetMap() map[string]interface{} {
	bag := baggage.FromContext(b.ctx)
	result := make(map[string]interface{})
	for _, member := range bag.Members() {
		result[member.Key()] = member.Value()
	}
	return result
}

// GetVar gets the baggage value for the specified key.
func (b *Baggage) GetVar(key string) *mvar.Var {
	bag := baggage.FromContext(b.ctx)
	member := bag.Member(key)
	return mvar.New(member.Value())
}
