package v1

import "github.com/mingzaily/maltose/frame/m"

type HelloReq struct {
	m.Meta `method:"GET" path:"/hello" summary:"Hello请求" tag:"公共服务" dc:"AAAAABBBBBB"`
	Name   string `form:"name" dc:"姓名" binding:"required"`
	Pass   string `form:"pass" dc:"密码" binding:"required"`
}

type HelloRes struct {
	Name string `json:"name" dc:"姓名"`
}

type ByeReq struct {
	m.Meta `method:"POST" path:"/bye" summary:"Bye请求" tag:"公共服务" dc:"BBBBBBCCCCCCC"`
	Name   string `json:"name" dc:"姓名" binding:"required"`
}

type ByeRes struct {
	Name string `json:"name" dc:"姓名"`
}
