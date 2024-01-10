package httpraw

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	"github.com/suifengpiao14/httpraw/thirdlib"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
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
	RawTpl            string `json:"rawTpl"`
	DynamicScript     string `json:"dynamicScript"`
	httpTpl           *httpTpl
	curlHook          CURLHookI
	GetCURLHookImplFn string `json:"getCURLHookImplFn"`
}

func NewHttpProxy(rawTpl string, dynamicScript string, getCURLHookImplFn string) (httpProxy *HttpProxy, err error) {
	httpTpl, err := NewHttpTpl(rawTpl)
	if err != nil {
		return nil, err
	}

	httpProxy = &HttpProxy{
		DynamicScript:     dynamicScript,
		RawTpl:            rawTpl,
		httpTpl:           httpTpl,
		GetCURLHookImplFn: getCURLHookImplFn,
	}
	if dynamicScript == "" {
		return httpProxy, nil
	}
	if getCURLHookImplFn == "" {
		err = errors.Errorf("get CURLHook implement from dynamic script path required when dynimicScrip is: %s", dynamicScript)
		return nil, err
	}
	// 解析动态脚本
	interpreter := interp.New(interp.Options{})
	interpreter.Use(stdlib.Symbols)
	interpreter.Use(thirdlib.Symbols) // 注册第三方库
	interpreter.Use(Symbols)          //注册当前包结构体

	_, err = interpreter.Eval(dynamicScript)
	if err != nil {
		err = errors.WithMessage(err, "init dynamic go script error")
		return nil, err
	}

	hookImpl, err := interpreter.Eval(httpProxy.GetCURLHookImplFn)
	if err != nil {
		err = errors.WithMessage(err, "selector fomat is packageName.getCurlHookImplFnName, for example curlhook.NewCURLHook . curlhook is dynamic script package name ,NewCURLHook is a func defined func()httpraw.CURLHookI")
		return nil, err
	}
	interfa := hookImpl.Interface()
	curlHookImplFn, ok := interfa.(func() CURLHookI)
	if !ok {
		err = errors.Errorf("dynamic func %s ,must return CURLHookI implement", httpProxy.GetCURLHookImplFn)
		return nil, err
	}

	httpProxy.curlHook = curlHookImplFn()

	return httpProxy, nil
}

//Request 发起请求，data 是tpl中用到的数据，scriptData 是动态脚本内全局变量
func (hp HttpProxy) Request(data any, scriptData map[string]interface{}, transport *http.Transport) (out []byte, err error) {
	r, err := hp.httpTpl.Request(data)
	if err != nil {
		return nil, err
	}
	if hp.curlHook != nil {
		rDTO, err := DestructReqeust(r)
		if err != nil {
			return nil, err
		}
		newRDTO, err := hp.curlHook.BeforeFn(*rDTO, scriptData)
		if err != nil {
			return nil, err
		}
		r, err = BuildRequest(newRDTO)
		if err != nil {
			return nil, err
		}
	}

	ctx := context.Background()
	out, err = RestyRequestFn(ctx, r, transport)
	if err != nil {
		return nil, err
	}
	if hp.curlHook != nil {
		out, err = hp.curlHook.AfterFn(out, scriptData)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
