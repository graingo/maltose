package mclient

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/graingo/maltose/errors/merror"
)

// ClientConfig is the configuration for Client.
type ClientConfig struct {
	// BaseURL specifies the base URL for all requests.
	BaseURL string `mconv:"base_url"`
	// Timeout specifies a time limit for requests made by this client.
	Timeout time.Duration `mconv:"timeout"`
	// Transport specifies the mechanism by which individual HTTP requests are made.
	Transport http.RoundTripper
	// Header specifies the default header for requests.
	Header http.Header
}

// SetBrowserMode enables browser mode of the client.
// When browser mode is enabled, it automatically saves and sends cookie content
// from and to server via an in-memory cookie jar.
func (c *Client) SetBrowserMode(enabled bool) *Client {
	if enabled {
		jar, _ := cookiejar.New(nil)
		c.client.Jar = jar
	}
	return c
}

// SetHeader sets a custom HTTP header pair for the client.
func (c *Client) SetHeader(key, value string) *Client {
	if c.config.Header == nil {
		c.config.Header = make(http.Header)
	}
	c.config.Header.Set(key, value)
	return c
}

// SetHeaderMap sets custom HTTP headers with map.
func (c *Client) SetHeaderMap(m map[string]string) *Client {
	if c.config.Header == nil {
		c.config.Header = make(http.Header)
	}
	for k, v := range m {
		c.config.Header.Set(k, v)
	}
	return c
}

// SetAgent sets the User-Agent header for client.
func (c *Client) SetAgent(agent string) *Client {
	return c.SetHeader("User-Agent", agent)
}

// SetContentType sets HTTP content type for the client.
func (c *Client) SetContentType(contentType string) *Client {
	return c.SetHeader("Content-Type", contentType)
}

// SetCookie sets a cookie pair for the client.
func (c *Client) SetCookie(key, value string) *Client {
	if c.client.Jar == nil {
		c.SetBrowserMode(true)
	}
	// Set cookie through header for now
	// In a real implementation, you would use the jar properly
	return c.SetHeader("Cookie", key+"="+value)
}

// SetCookieMap sets cookie items with map.
func (c *Client) SetCookieMap(m map[string]string) *Client {
	if c.client.Jar == nil {
		c.SetBrowserMode(true)
	}
	// Set cookies through header for now
	for k, v := range m {
		c.SetCookie(k, v)
	}
	return c
}

// SetBaseURL sets the base URL for all requests.
func (c *Client) SetBaseURL(baseURL string) *Client {
	c.config.BaseURL = baseURL
	return c
}

// SetTimeout sets the request timeout for the client.
func (c *Client) SetTimeout(t time.Duration) *Client {
	c.client.Timeout = t
	c.config.Timeout = t
	return c
}

// SetRedirectLimit limits the number of jumps for redirection.
// It sets the CheckRedirect function on the underlying http.Client.
// A limit of 0 means no redirects will be followed.
func (c *Client) SetRedirectLimit(redirectLimit int) *Client {
	c.client.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		if len(via) >= redirectLimit {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return c
}

// SetTLSKeyCrt sets client TLS certificate and key files.
func (c *Client) SetTLSKeyCrt(crtFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		return merror.Wrapf(err, "failed to load certificate from %s and key from %s", crtFile, keyFile)
	}

	if transport, ok := c.client.Transport.(*http.Transport); ok {
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		transport.TLSClientConfig = tlsConfig
		return nil
	}
	return merror.New("cannot set TLSClientConfig for custom Transport of the client")
}

// SetTLSConfig sets the client's TLS configuration.
func (c *Client) SetTLSConfig(tlsConfig *tls.Config) error {
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = tlsConfig
		return nil
	}
	return merror.New("cannot set TLSClientConfig for custom Transport of the client")
}

// SetBasicAuth sets HTTP basic authentication for the client.
func (c *Client) SetBasicAuth(username, password string) *Client {
	auth := username + ":" + password
	return c.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
}
