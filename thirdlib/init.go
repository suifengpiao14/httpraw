package thirdlib

import (
	_ "github.com/tidwall/gjson"
	_ "github.com/tidwall/sjson"
	"github.com/traefik/yaegi/stdlib"
)

// Symbols variable stores the map of stdlib symbols per package.
var Symbols = stdlib.Symbols

//go:generate go install github.com/traefik/yaegi/cmd/yaegi
//go:generate yaegi extract github.com/tidwall/gjson
//go:generate yaegi extract github.com/tidwall/sjson
