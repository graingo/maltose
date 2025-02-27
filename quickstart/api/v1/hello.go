package v1

import "github.com/mingzaily/maltose/frame/m"

type HelloReq struct {
	m.Meta `method:"GET" path:"/hello" summary:"Hello请求" tag:"公共服务"`
	Name   string `form:"name" dc:"姓名"`
}

type HelloRes struct {
	Name string `json:"name"`
}

type ByeReq struct {
	m.Meta `method:"POST" path:"/bye" summary:"Bye请求" tag:"公共服务"`
	Name   string `json:"name" dc:"姓名"`
}

type ByeRes struct {
	Name string `json:"name"`
}
