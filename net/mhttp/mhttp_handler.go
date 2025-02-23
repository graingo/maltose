package mhttp

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mingzaily/maltose/util/mmeta"
)

// HandlerMeta 处理器元数据
type HandlerMeta struct {
	Path        string
	Method      string
	Summary     string
	Description string
	Request     interface{}
	Response    interface{}
	Handler     gin.HandlerFunc
	Middleware  []gin.HandlerFunc
	Tags        []string
}

// MetaHandler 元数据处理器接口
type MetaHandler interface {
	Meta() *HandlerMeta
	Handle(*Request) (interface{}, error)
}

// HandlerFunc 定义基础处理函数类型
type HandlerFunc func(*Request)

// BindHandler 绑定简单处理函数
func (s *Server) BindHandler(method, path string, handler HandlerFunc) {
	s.Handle(method, path, func(c *gin.Context) {
		handler(&Request{Context: c})
	})
}

// Bind 绑定对象到路由
func (s *Server) Bind(object interface{}) {
	// 获取对象的类型和值
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

		// 注册路由处理函数
		s.Handle(httpMethod, path, func(c *gin.Context) {
			r := &Request{Context: c}
			req := reflect.New(reqElem).Interface()

			if err := handleRequest(r, method, val, req); err != nil {
				c.Error(err)
				return
			}
		})
	}
}

// handleRequest 处理请求并返回结果
func handleRequest(r *Request, method reflect.Method, val reflect.Value, req interface{}) error {
	// 参数绑定
	if err := r.ShouldBind(req); err != nil {
		return err
	}

	// 调用方法
	results := method.Func.Call([]reflect.Value{
		val,
		reflect.ValueOf(r.Request.Context()),
		reflect.ValueOf(req),
	})

	// 处理返回值
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	// 设置响应到 Request 中
	r.SetHandlerResponse(results[0].Interface())
	return nil
}
