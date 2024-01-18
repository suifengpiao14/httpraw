package curlhookyaegi

import (
	"github.com/pkg/errors"
	_ "github.com/spf13/cast"
	"github.com/suifengpiao14/httpraw"
	_ "github.com/tidwall/gjson"
	_ "github.com/tidwall/sjson"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var Symbols = stdlib.Symbols
var CURLHookImplPoint = "curlhook.NewCURLHook"

func NewCurlHookYaegi(dynamicScript string) (curlHook httpraw.CURLHookI, err error) {

	// 解析动态脚本
	interpreter := interp.New(interp.Options{})
	interpreter.Use(stdlib.Symbols)

	interpreter.Use(Symbols) //注册当前包结构体

	_, err = interpreter.Eval(dynamicScript)
	if err != nil {
		err = errors.WithMessage(err, "init dynamic go script error")
		return nil, err
	}

	hookImpl, err := interpreter.Eval(CURLHookImplPoint)
	if err != nil {
		err = errors.WithMessage(err, "dynamic script packge must curlhook and have function func NewCURLHook()httpraw.CURLHookI")
		return nil, err
	}
	interfa := hookImpl.Interface()
	curlHookImplFn, ok := interfa.(func() httpraw.CURLHookI)
	if !ok {
		err = errors.Errorf("dynamic func %s ,must return CURLHookI implement", CURLHookImplPoint)
		return nil, err
	}
	curlHook = curlHookImplFn()
	return curlHook, nil
}

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/suifengpiao14/httpraw
//go:generate yaegi extract github.com/tidwall/gjson
//go:generate yaegi extract github.com/tidwall/sjson
//go:generate yaegi extract github.com/spf13/cast
