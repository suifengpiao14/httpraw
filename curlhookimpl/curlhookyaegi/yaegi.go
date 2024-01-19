package curlhookyaegi

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	_ "github.com/spf13/cast"
	"github.com/suifengpiao14/httpraw"
	_ "github.com/tidwall/gjson"
	_ "github.com/tidwall/sjson"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var Symbols = stdlib.Symbols
var CURLHookBeforeFnPoint = "curlhook.BeforeFn"
var CURLHookAfterFnPoint = "curlhook.AfterFn"

const (
	undefined_selector_error_prefix = "undefined selector: "
)

//ValidateDynamicScript  用于往数据库预先写入动态脚本时验证合法性
func ValidateDynamicScript(dynamicScript string) (err error) {
	_, err = NewCurlHookYaegi(dynamicScript)
	return err
}

func NewCurlHookYaegi(dynamicScript string) (curlHook httpraw.CURLHookI, err error) {
	var (
		beforeFn       httpraw.BeforeFn
		afterFn        httpraw.AfterFn
		beforeFnExists = true
		afterFnExists  = true
	)

	// 解析动态脚本
	interpreter := interp.New(interp.Options{})
	interpreter.Use(stdlib.Symbols)

	interpreter.Use(Symbols) //注册当前包结构体

	_, err = interpreter.Eval(dynamicScript)
	if err != nil {
		err = errors.WithMessage(err, "init dynamic go script error")
		return nil, err
	}

	beforeFnT := reflect.TypeOf((httpraw.BeforeFn)(nil))
	beforeFnV, beforeFnExists, err := getFn(interpreter, CURLHookBeforeFnPoint, beforeFnT)
	if err != nil {
		return nil, err
	}

	if beforeFnExists {
		beforeFn = beforeFnV.Interface().(httpraw.BeforeFn)
	}
	afterFnT := reflect.TypeOf((httpraw.AfterFn)(nil))
	afterFnV, afterFnExists, err := getFn(interpreter, CURLHookAfterFnPoint, afterFnT)
	if err != nil {
		return nil, err
	}
	if afterFnExists {
		afterFn = afterFnV.Interface().(httpraw.AfterFn)
	}
	curlHook = httpraw.NewDynamicCURLHook(beforeFn, afterFn)
	return curlHook, nil
}

//getFn 从动态脚本中获取特定函数
func getFn(interpreter *interp.Interpreter, selector string, dstType reflect.Type) (fn reflect.Value, exists bool, err error) {
	fnV, err := interpreter.Eval(selector)
	if err != nil && strings.Contains(err.Error(), undefined_selector_error_prefix) { // 不存在当前元素 时 忽略错误，程序容许只动态实现一部分
		err = nil
		return fn, false, nil
	}

	if err != nil {
		err = errors.WithMessage(err, selector)
		return fn, false, err
	}
	if !fnV.CanConvert(dstType) {
		err = errors.Errorf("dynamic func %s ,must can convert to %s", selector, dstType.PkgPath())
		return fn, true, err
	}
	fn = fnV.Convert(dstType)
	return fn, true, nil
}

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/suifengpiao14/httpraw
//go:generate yaegi extract github.com/tidwall/gjson
//go:generate yaegi extract github.com/tidwall/sjson
//go:generate yaegi extract github.com/spf13/cast
