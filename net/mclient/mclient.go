package mclient

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/graingo/maltose"
)

// Client 是 HTTP 客户端，用于管理 HTTP 请求。
type Client struct {
	http.Client                         // 底层 HTTP 客户端。
	baseURL           string            // 请求的基础 URL。
	header            map[string]string // 自定义请求头。
	cookies           map[string]string // 自定义 Cookie。
	retryCount        int               // 请求失败重试次数。
	retryInterval     time.Duration     // 请求失败重试间隔。
	middlewareHandler []HandlerFunc     // 拦截器处理程序。
}

const (
	httpProtocolName          = `http`
	httpParamFileHolder       = `@file:`
	httpRegexParamJson        = `^[\w\[\]]+=.+`
	httpRegexHeaderRaw        = `^([\w\-]+):\s*(.+)`
	httpHeaderHost            = `Host`
	httpHeaderCookie          = `Cookie`
	httpHeaderUserAgent       = `User-Agent`
	httpHeaderContentType     = `Content-Type`
	httpHeaderContentTypeJson = `application/json`
	httpHeaderContentTypeXml  = `application/xml`
	httpHeaderContentTypeForm = `application/x-www-form-urlencoded`
)

var (
	hostname, _      = os.Hostname()
	defaultUserAgent = fmt.Sprintf(`mclient %s`, maltose.VERSION)
)

// New 创建一个新的 HTTP 客户端
func New() *Client {
	c := &Client{
		Client: http.Client{
			Transport: &http.Transport{
				// 默认情况下，不验证 HTTPS 证书
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				DisableKeepAlives: true,
			},
		},
		header:  make(map[string]string),
		cookies: make(map[string]string),
	}
	c.header[httpHeaderUserAgent] = defaultUserAgent
	return c
}

// Clone 克隆当前客户端
func (c *Client) Clone() *Client {
	newClient := New()
	*newClient = *c
	newClient.header = make(map[string]string, len(c.header))
	for k, v := range c.header {
		newClient.header[k] = v
	}
	newClient.cookies = make(map[string]string, len(c.cookies))
	for k, v := range c.cookies {
		newClient.cookies[k] = v
	}
	return newClient
}
