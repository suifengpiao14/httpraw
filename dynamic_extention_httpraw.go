package httpraw

import (
	"github.com/suifengpiao14/yaegijson"
)

var Symbols = yaegijson.Symbols

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/suifengpiao14/httpraw

func NewDynamicExtensionHttpRaw(extensionCode string, extensionPath string) *yaegijson.DynamicExtension {
	extension := yaegijson.NewDynamicExtension(extensionCode, extensionPath).Withsymbols(Symbols)
	return extension
}
