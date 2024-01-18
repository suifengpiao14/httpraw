package httpraw

import (
	"context"
	"net/http"
)

// http 代理请求
type CURLHookI interface {
	BeforeFn(r RequestDTO, scriptData map[string]interface{}) (nr *RequestDTO, err error) // 请求前钩子函数,scriptData 用来传递请求时的额外数据，如循环请求时，循环次数
	AfterFn(body []byte, scriptData map[string]interface{}) (newBody []byte, err error)   // 请求后钩子函数,scriptData 用来传递请求时的额外数据，如循环请求时，循环次数
}

type EmptyCURLHookImpl struct{}

func (impl EmptyCURLHookImpl) BeforeFn(r RequestDTO, scriptData map[string]interface{}) (nr *RequestDTO, err error) {
	return &r, nil
}
func (impl EmptyCURLHookImpl) AfterFn(body []byte, scriptData map[string]interface{}) (newBody []byte, err error) {
	return body, nil
}

type HttpProxy struct {
	RawTpl string `json:"rawTpl"`

	httpTpl  *httpTpl
	curlHook CURLHookI
}

func NewHttpProxy(rawTpl string, curlHook CURLHookI) (httpProxy *HttpProxy, err error) {
	httpTpl, err := NewHttpTpl(rawTpl)
	if err != nil {
		return nil, err
	}

	httpProxy = &HttpProxy{
		RawTpl:   rawTpl,
		httpTpl:  httpTpl,
		curlHook: curlHook,
	}

	return httpProxy, nil
}

// RequestDTO 数据转RequestDTO 格式
func (hp HttpProxy) RequestDTO(data any) (rDTO *RequestDTO, err error) {
	r, err := hp.httpTpl.Request(data)
	if err != nil {
		return nil, err
	}
	rDTO, err = DestructReqeust(r)
	if err != nil {
		return nil, err
	}
	return rDTO, nil
}

// Request 发起请求，data 是tpl中用到的数据，scriptData 是动态脚本内全局变量
func (hp HttpProxy) Request(rDTO *RequestDTO, scriptData map[string]any, transport *http.Transport) (reqDTo *RequestDTO, out []byte, err error) {
	reqDTo = rDTO
	if hp.curlHook != nil {
		reqDTo, err = hp.curlHook.BeforeFn(*rDTO, scriptData) //修改请求参数
		if err != nil {
			return nil, nil, err
		}
	}
	r, err := BuildRequest(reqDTo)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	out, err = RestyRequestFn(ctx, r, transport)
	if err != nil {
		return nil, nil, err
	}
	if hp.curlHook != nil {
		out, err = hp.curlHook.AfterFn(out, scriptData)
		if err != nil {
			return nil, nil, err
		}
	}
	return reqDTo, out, nil
}
