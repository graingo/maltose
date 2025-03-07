package mtrace

import (
	"context"

	"github.com/graingo/maltose/container/mvar"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/baggage"
)

// Baggage 是一种在分布式系统的上下文中传播键值对数据的机制
// 它允许将自定义数据（如用户ID、请求ID等）附加到跟踪中，并在服务调用之间传播
type Baggage struct {
	ctx context.Context
}

// NewBaggage 创建一个新的 Baggage 实例
func NewBaggage(ctx context.Context) *Baggage {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Baggage{
		ctx: ctx,
	}
}

// SetValue 设置单个 baggage 值
func (b *Baggage) SetValue(key string, value interface{}) context.Context {
	member, _ := baggage.NewMember(key, cast.ToString(value))
	bag, _ := baggage.New(member)
	b.ctx = baggage.ContextWithBaggage(b.ctx, bag)
	return b.ctx
}

// SetMap 批量设置 baggage 值
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

// GetMap 获取所有 baggage 值
func (b *Baggage) GetMap() map[string]interface{} {
	bag := baggage.FromContext(b.ctx)
	result := make(map[string]interface{})
	for _, member := range bag.Members() {
		result[member.Key()] = member.Value()
	}
	return result
}

// GetVar 获取指定 key 的 baggage 值
func (b *Baggage) GetVar(key string) *mvar.Var {
	bag := baggage.FromContext(b.ctx)
	member := bag.Member(key)
	return mvar.New(member.Value())
}
