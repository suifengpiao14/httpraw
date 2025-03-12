package httpraw

import (
	"github.com/suifengpiao14/yaegijson"
)

var Symbols = yaegijson.Symbols

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/suifengpiao14/httpraw

type DynamicHook struct {
	BeforeRequestFuncName string `json:"beforeRequestFuncName"`
	AfterRequestFuncName  string `json:"afterRequestFuncName"`
	ExtensionCode         string `json:"extensionCode"`
	ExtensionPath         string `json:"extensionPath"`
}

func (p DynamicHook) HookFn() (beforeRequestFunc BeforRequestFn, afterRequestFunc AfterRequestFn, err error) {
	// 动态编译扩展代码
	extension := yaegijson.NewDynamicExtension(p.ExtensionCode, p.ExtensionPath).Withsymbols(Symbols)
	err = extension.GetDestFuncImpl(p.BeforeRequestFuncName, &beforeRequestFunc)
	if err != nil {
		return nil, nil, err
	}
	err = extension.GetDestFuncImpl(p.AfterRequestFuncName, &afterRequestFunc)
	if err != nil {
		return nil, nil, err
	}
	return beforeRequestFunc, afterRequestFunc, nil
}
