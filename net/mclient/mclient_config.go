package mclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// SetBrowserMode 启用客户端的浏览器模式。
// 当浏览器模式启用时，它会自动保存并发送 Cookie 内容到服务器和从服务器接收。
func (c *Client) SetBrowserMode(enabled bool) *Client {
	if enabled {
		jar, _ := cookiejar.New(nil)
		c.Jar = jar
	}
	return c
}

// SetHeader 为客户端设置自定义 HTTP 头部对。
func (c *Client) SetHeader(key, value string) *Client {
	c.header[key] = value
	return c
}

// SetHeaderMap 使用映射设置自定义 HTTP 头部。
func (c *Client) SetHeaderMap(m map[string]string) *Client {
	for k, v := range m {
		c.header[k] = v
	}
	return c
}

// SetAgent 设置客户端的 User-Agent 头部。
func (c *Client) SetAgent(agent string) *Client {
	c.header[httpHeaderUserAgent] = agent
	return c
}

// SetContentType 为客户端设置 HTTP 内容类型。
func (c *Client) SetContentType(contentType string) *Client {
	c.header[httpHeaderContentType] = contentType
	return c
}

// SetHeaderRaw 使用原始字符串设置自定义 HTTP 头部。
func (c *Client) SetHeaderRaw(headers string) *Client {
	for _, line := range gstr.SplitAndTrim(headers, "\n") {
		array, _ := gregex.MatchString(httpRegexHeaderRaw, line)
		if len(array) >= 3 {
			c.header[array[1]] = array[2]
		}
	}
	return c
}

// SetCookie 为客户端设置一个 Cookie 对。
func (c *Client) SetCookie(key, value string) *Client {
	c.cookies[key] = value
	return c
}

// SetCookieMap 使用映射设置 Cookie 项。
func (c *Client) SetCookieMap(m map[string]string) *Client {
	for k, v := range m {
		c.cookies[k] = v
	}
	return c
}

// SetPrefix 设置请求服务器 URL 前缀。
func (c *Client) SetPrefix(prefix string) *Client {
	c.prefix = prefix
	return c
}

// SetTimeout 为客户端设置请求超时时间。
func (c *Client) SetTimeout(t time.Duration) *Client {
	c.Client.Timeout = t
	return c
}

// SetBasicAuth 为客户端设置 HTTP 基本认证信息。
func (c *Client) SetBasicAuth(user, pass string) *Client {
	c.authUser = user
	c.authPass = pass
	return c
}

// SetRetry 设置重试次数和间隔。
// TODO 已移除。
func (c *Client) SetRetry(retryCount int, retryInterval time.Duration) *Client {
	c.retryCount = retryCount
	c.retryInterval = retryInterval
	return c
}

// SetRedirectLimit 限制跳转次数。
func (c *Client) SetRedirectLimit(redirectLimit int) *Client {
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= redirectLimit {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return c
}

// SetNoUrlEncode 设置标记，在发送请求前不对参数进行编码。
func (c *Client) SetNoUrlEncode(noUrlEncode bool) *Client {
	c.noUrlEncode = noUrlEncode
	return c
}

// SetProxy 为客户端设置代理。
// 当参数 `proxyURL` 为空或格式错误时，此函数不会执行任何操作。
// 正确的格式如 `http://用户:密码@IP:端口` 或 `socks5://用户:密码@IP:端口`。
// 目前仅支持 `http` 和 `socks5` 代理。
func (c *Client) SetProxy(proxyURL string) {
	if strings.TrimSpace(proxyURL) == "" {
		return
	}
	_proxy, err := url.Parse(proxyURL)
	if err != nil {
		intlog.Errorf(context.TODO(), `%+v`, err)
		return
	}
	if _proxy.Scheme == httpProtocolName {
		if v, ok := c.Transport.(*http.Transport); ok {
			v.Proxy = http.ProxyURL(_proxy)
		}
	} else {
		auth := &proxy.Auth{}
		user := _proxy.User.Username()
		if user != "" {
			auth.User = user
			password, hasPassword := _proxy.User.Password()
			if hasPassword && password != "" {
				auth.Password = password
			}
		} else {
			auth = nil
		}
		// 参考源代码，错误始终为 nil
		dialer, err := proxy.SOCKS5(
			"tcp",
			_proxy.Host,
			auth,
			&net.Dialer{
				Timeout:   c.Client.Timeout,
				KeepAlive: c.Client.Timeout,
			},
		)
		if err != nil {
			intlog.Errorf(context.TODO(), `%+v`, err)
			return
		}
		if v, ok := c.Transport.(*http.Transport); ok {
			v.DialContext = func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				return dialer.Dial(network, addr)
			}
		}
		// c.SetTimeout(10*time.Second)
	}
}

// SetTLSKeyCrt 为客户端的 TLS 配置设置证书和密钥文件。
func (c *Client) SetTLSKeyCrt(crtFile, keyFile string) error {
	tlsConfig, err := LoadKeyCrt(crtFile, keyFile)
	if err != nil {
		return gerror.Wrap(err, "LoadKeyCrt 失败")
	}
	if v, ok := c.Transport.(*http.Transport); ok {
		tlsConfig.InsecureSkipVerify = true
		v.TLSClientConfig = tlsConfig
		return nil
	}
	return gerror.New(`无法为客户端的自定义 Transport 设置 TLSClientConfig`)
}

// SetTLSConfig 设置客户端的 TLS 配置。
func (c *Client) SetTLSConfig(tlsConfig *tls.Config) error {
	if v, ok := c.Transport.(*http.Transport); ok {
		v.TLSClientConfig = tlsConfig
		return nil
	}
	return gerror.New(`无法为客户端的自定义 Transport 设置 TLSClientConfig`)
}
