package mtrace

import (
	"context"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/mconv"
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
	member, _ := baggage.NewMember(key, mconv.ToString(value))
	// Correctly create a new baggage with the new member.
	// We must start from the existing baggage in the context.
	bag := baggage.FromContext(b.ctx)
	bag, _ = bag.SetMember(member)
	b.ctx = baggage.ContextWithBaggage(b.ctx, bag)
	return b.ctx
}

// SetMap sets multiple baggage values.
func (b *Baggage) SetMap(data map[string]interface{}) context.Context {
	bag := baggage.FromContext(b.ctx)
	for k, v := range data {
		member, _ := baggage.NewMember(k, mconv.ToString(v))
		bag, _ = bag.SetMember(member)
	}
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
	member := baggage.FromContext(b.ctx).Member(key)
	// If the member does not exist, the value is empty,
	// but we should return a nil-value Var.
	if member.Key() == "" {
		return mvar.New(nil)
	}
	return mvar.New(member.Value())
}
