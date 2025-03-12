package httpraw_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/httpraw"
)

var extensionCode = `
package dynamichookexample

import (
	"github.com/suifengpiao14/httpraw"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func BeforeFn(input *httpraw.RequestDTO) (output *httpraw.RequestDTO, err error) {
	timestamps := gjson.Get(input.Body, "body._head._timestamps").String()
	_ = timestamps
	input.Body, err = sjson.Set(input.Body, "body._head._timestamps", "1111111111111111")
	if err != nil {
		return nil, err
	}
	return input, nil
}
func AfterFn(input *httpraw.ResponseDTO) (output *httpraw.ResponseDTO, err error) {
	return input, nil
}

`

func TestDynamicHook(t *testing.T) {
	dynamicExtensionHttpRaw := httpraw.NewDynamicExtensionHttpRaw(extensionCode, "")
	dynamicHook := httpraw.DynamicHook{
		BeforeRequestFuncName:   "dynamichookexample.BeforeFn",
		AfterRequestFuncName:    "dynamichookexample.AfterFn",
		DynamicExtensionHttpRaw: dynamicExtensionHttpRaw,
	}

	before, after, _, err := dynamicHook.HookFn()
	require.NoError(t, err)
	require.NotNil(t, before)
	require.NotNil(t, after)

	input := &httpraw.RequestDTO{
		Body: `{"body":{"_head":{"_timestamps":"1234567890"}}}`,
	}
	input, err = before(input)
	require.NoError(t, err)
	require.JSONEq(t, `{"body":{"_head":{"_timestamps":"1111111111111111"}}}`, input.Body)

}
