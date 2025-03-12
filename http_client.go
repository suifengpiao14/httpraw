package httpraw

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type Client struct {
	TransportConfig *TransportConfig
	_Transport      *http.Transport
}

func NewClient(transportConfig *TransportConfig) Client {
	return Client{
		TransportConfig: transportConfig,
	}
}

func (c Client) Transport() *http.Transport {
	if c._Transport != nil {
		return c._Transport
	}
	if c.TransportConfig == nil {
		c._Transport = http.DefaultTransport.(*http.Transport)
		return c._Transport
	}
	c._Transport = NewTransport(c.TransportConfig)
	return c._Transport
}

func (c Client) Execute(ctx context.Context, req *http.Request) (out []byte, resp *http.Response, err error) {
	client := resty.New()
	client.SetTransport(c.Transport())
	r := client.R().SetContext(ctx)
	urlstr := req.URL.String()
	r.Header = req.Header
	r.FormData = req.Form
	r.RawRequest = req
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, nil, err
		}
		r.SetBody(body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	res, err := r.Execute(strings.ToUpper(req.Method), urlstr)
	if err != nil {
		return nil, nil, err
	}
	responseBody := res.Body()
	if !res.IsSuccess() {
		err = errors.Errorf("http status:%d,body:%s", res.StatusCode(), string(responseBody))
		err = errors.WithMessage(err, fmt.Sprintf("%v", res.Error()))
		return nil, nil, err
	}
	return responseBody, res.RawResponse, nil

}

var CURL_TIMEOUT = 30 * time.Millisecond

type TransportConfig struct {
	Proxy               string `json:"proxy"`
	Timeout             int    `json:"timeout"`
	KeepAlive           int    `json:"keepAlive"`
	MaxIdleConns        int    `json:"maxIdleConns"`
	MaxIdleConnsPerHost int    `json:"maxIdleConnsPerHost"`
	IdleConnTimeout     int    `json:"idleConnTimeout"`
}

// NewTransport 创建一个htt连接,兼容代理模式
func NewTransport(cfg *TransportConfig) *http.Transport {
	maxIdleConns := 200
	maxIdleConnsPerHost := 20
	idleConnTimeout := 90
	if cfg.MaxIdleConns > 0 {
		maxIdleConns = cfg.MaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost > 0 {
		maxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	}
	if cfg.IdleConnTimeout > 0 {
		idleConnTimeout = cfg.IdleConnTimeout
	}
	timeout := 10
	if cfg.Timeout > 0 {
		timeout = 10
	}
	keepAlive := 300
	if cfg.KeepAlive > 0 {
		keepAlive = cfg.KeepAlive
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,   // 连接超时时间
			KeepAlive: time.Duration(keepAlive) * time.Second, // 连接保持超时时间
		}).DialContext,
		MaxIdleConns:        maxIdleConns,                                 // 最大连接数,默认0无穷大
		MaxIdleConnsPerHost: maxIdleConnsPerHost,                          // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
		IdleConnTimeout:     time.Duration(idleConnTimeout) * time.Second, // 多长时间未使用自动关闭连
	}
	if cfg.Proxy != "" {
		proxy, err := url.Parse(cfg.Proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	return transport
}
