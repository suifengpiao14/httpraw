package curlhookyaegi

import (
	"reflect"

	"github.com/suifengpiao14/httpraw"
)

func init() {
	Symbols["github.com/suifengpiao14/httpraw/httpraw"] = map[string]reflect.Value{
		// type definitions
		"RequestDTO": reflect.ValueOf((*httpraw.RequestDTO)(nil)),
	}
}
