package mhttp

// HandlerFunc 定义基础处理函数类型
type HandlerFunc func(*Request)

// BindHandler 绑定简单处理函数
func (s *Server) BindHandler(method, path string, handler HandlerFunc) {
	// 保存路由信息
	s.routes = append(s.routes, Route{
		Method:      method,
		Path:        path,
		HandlerFunc: handler,
		Type:        routeTypeHandler,
	})

	// 添加到预绑定列表
	s.preBindItems = append(s.preBindItems, preBindItem{
		Method:      method,
		Path:        path,
		HandlerFunc: handler,
		Type:        routeTypeHandler,
	})
}

// Bind 绑定对象到根路由
func (s *Server) Bind(object any) {
	s.bindObject(nil, object)
}
