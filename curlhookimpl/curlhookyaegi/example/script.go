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
