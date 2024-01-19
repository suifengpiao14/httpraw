package curlhookyaegi_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/httpraw/curlhookimpl/curlhookyaegi"
)

func TestNewCurlHookYaegi(t *testing.T) {
	t.Run("has BeforeFn and AfterFn", func(t *testing.T) {
		dynamic := `
		package curlhook
	
	import (
		"github.com/suifengpiao14/httpraw"
		"github.com/tidwall/gjson"
		"github.com/tidwall/sjson"
	)
	
	func BeforeFn(r httpraw.RequestDTO, scriptData map[string]interface{}) (nr *httpraw.RequestDTO, err error) {
		timestamps := gjson.Get(r.Body, "_head._timestamps").String()
		_ = timestamps
		r.Body, err = sjson.Set(r.Body, "_head._timestamps", "1111111111111111")
		if err != nil {
			return nil, err
		}
		return &r, nil
	
	}
	func AfterFn(body []byte, scriptData map[string]interface{}) (newBody []byte, err error) {
		return body, nil
	}
	
		`
		_, err := curlhookyaegi.NewCurlHookYaegi(dynamic)
		require.NoError(t, err)
	})

	t.Run("only BeforFn", func(t *testing.T) {
		dynamic := `
		package curlhook
	
	import (
		"github.com/suifengpiao14/httpraw"
		"github.com/tidwall/gjson"
		"github.com/tidwall/sjson"
	)
	
	func BeforeFn(r httpraw.RequestDTO, scriptData map[string]interface{}) (nr *httpraw.RequestDTO, err error) {
		timestamps := gjson.Get(r.Body, "_head._timestamps").String()
		_ = timestamps
		r.Body, err = sjson.Set(r.Body, "_head._timestamps", "1111111111111111")
		if err != nil {
			return nil, err
		}
		return &r, nil
	
	}
		`
		_, err := curlhookyaegi.NewCurlHookYaegi(dynamic)
		require.NoError(t, err)
	})

	t.Run("only AfterFn", func(t *testing.T) {
		dynamic := `
		package curlhook
	
	func AfterFn(body []byte, scriptData map[string]interface{}) (newBody []byte, err error) {
		return body, nil
	}
		`
		_, err := curlhookyaegi.NewCurlHookYaegi(dynamic)
		require.NoError(t, err)
	})

}
