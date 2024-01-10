// Code generated by 'yaegi extract github.com/suifengpiao14/httpraw'. DO NOT EDIT.

package httpraw

import (
	"reflect"
)

func init() {
	Symbols["github.com/suifengpiao14/httpraw/httpraw"] = map[string]reflect.Value{
		// function, constant and variable definitions
		// type definitions
		"CURLHookI":         reflect.ValueOf((*CURLHookI)(nil)),
		"EmptyCURLHookImpl": reflect.ValueOf((*EmptyCURLHookImpl)(nil)),
		"RequestDTO":        reflect.ValueOf((*RequestDTO)(nil)),

		// interface wrapper definitions
		"_CURLHookI": reflect.ValueOf((*_github_com_suifengpiao14_httpraw_CURLHookI)(nil)),
	}
}

// _github_com_suifengpiao14_httpraw_CURLHookI is an interface wrapper for CURLHookI type
type _github_com_suifengpiao14_httpraw_CURLHookI struct {
	IValue    interface{}
	WAfterFn  func(body []byte, scriptData map[string]interface{}) (newBody []byte, err error)
	WBeforeFn func(r RequestDTO, scriptData map[string]interface{}) (nr *RequestDTO, err error)
}

func (W _github_com_suifengpiao14_httpraw_CURLHookI) AfterFn(body []byte, scriptData map[string]interface{}) (newBody []byte, err error) {
	return W.WAfterFn(body, scriptData)
}
func (W _github_com_suifengpiao14_httpraw_CURLHookI) BeforeFn(r RequestDTO, scriptData map[string]interface{}) (nr *RequestDTO, err error) {
	return W.WBeforeFn(r, scriptData)
}
