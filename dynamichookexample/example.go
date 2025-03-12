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
