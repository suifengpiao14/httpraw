package httpraw

import (
	"github.com/pkg/errors"
	"github.com/suifengpiao14/yaegijson"
)

var Symbols = yaegijson.Symbols

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/suifengpiao14/httpraw

type DynamicHook struct {
	BeforeRequestFuncName   string                      `json:"beforeRequestFuncName"`
	AfterRequestFuncName    string                      `json:"afterRequestFuncName"`
	DynamicExtensionHttpRaw *yaegijson.DynamicExtension `json:"-"`
}

func (p DynamicHook) HookFn() (beforeRequestFunc BeforeRequestFn, afterRequestFunc AfterRequestFn, err error) {
	// 动态编译扩展代码
	extension := p.DynamicExtensionHttpRaw
	if extension == nil {
		err = errors.Errorf("DynamicExtensionHttpRaw is nil")
		return nil, nil, err

	}
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

func NewDynamicExtensionHttpRaw(extensionCode string, extensionPath string) *yaegijson.DynamicExtension {
	extension := yaegijson.NewDynamicExtension(extensionCode, extensionPath).Withsymbols(Symbols)
	return extension
}
