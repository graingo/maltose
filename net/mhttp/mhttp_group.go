package mhttp

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mingzaily/maltose/util/mmeta"
)

// GroupHandler 定义路由组处理函数类型
type GroupHandler func(group *RouterGroup)

// RouterGroup 路由组
type RouterGroup struct {
	*gin.RouterGroup
	server *Server
}

// Group 创建路由组
func (s *Server) Group(prefix string, handlers ...interface{}) *RouterGroup {
	group := &RouterGroup{
		RouterGroup: s.Engine.Group(prefix),
		server:      s,
	}

	// 处理中间件
	for _, handler := range handlers {
		switch h := handler.(type) {
		case gin.HandlerFunc:
			group.Use(h)
		case GroupHandler:
			h(group)
		}
	}

	return group
}

// Bind 绑定控制器到路由组
func (g *RouterGroup) Bind(objects ...interface{}) *RouterGroup {
	for _, object := range objects {
		// 获取原始路径前缀
		prefix := g.BasePath()

		// 绑定到服务器
		g.server.bindToGroup(prefix, object)
	}
	return g
}

// Middleware 为路由组添加中间件
func (g *RouterGroup) Middleware(handlers ...gin.HandlerFunc) *RouterGroup {
	g.Use(handlers...)
	return g
}

// bindToGroup 将对象绑定到指定路由组
func (s *Server) bindToGroup(prefix string, object interface{}) {
	typ := reflect.TypeOf(object)
	val := reflect.ValueOf(object)

	// 遍历所有方法
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)

		// 检查方法签名是否符合要求: func(context.Context, *XxxReq) (*XxxRes, error)
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 2 {
			continue
		}

		// 获取请求参数类型
		reqType := method.Type.In(2)
		if reqType.Kind() != reflect.Ptr {
			continue
		}

		// 获取请求结构体的元数据
		reqElem := reqType.Elem()
		reqInstance := reflect.New(reqElem).Interface()

		// 查找嵌入的 Meta 字段
		metaField := reflect.ValueOf(reqInstance).Elem().FieldByName("Meta")
		if !metaField.IsValid() {
			continue
		}

		// 获取路由元数据
		path := mmeta.Get(reqInstance, "path").String()
		httpMethod := mmeta.Get(reqInstance, "method").String()

		if path == "" || httpMethod == "" {
			continue
		}

		// 组合完整路径
		fullPath := path
		if prefix != "/" {
			fullPath = prefix + path
		}

		// 注册路由处理函数
		s.Handle(httpMethod, fullPath, func(c *gin.Context) {
			r := &Request{Context: c}
			req := reflect.New(reqElem).Interface()

			if err := handleRequest(r, method, val, req); err != nil {
				c.Error(err)
				return
			}
		})
	}
}
