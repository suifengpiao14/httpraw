package httpraw_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/httpraw"
	"github.com/tidwall/gjson"
)

func TestCurlProxy(t *testing.T) {
	var httpTpl httpraw.HttpTpl = `
	POST / HTTP/1.1
	Host: new-merchant-api.hsb.com
	Content-Type: application/json
	HSB-OPENAPI-CALLERSERVICEID: 214001
	HSB-OPENAPI-SIGNATURE: 767a9cd8148fc5bc460c16372fbac532



	{"_head":{"_interface":"NewMerchantCenterServer.Api.V1.getMerchantInfo","_msgType":"request","_remark":"","_version":"0.01","_timestamps":"1439261904","_invokeId":"563447634257324435","_callerServiceId":"210015","_groupNo":"1"},"_param":{"merchantId":"{{merchantId}}","queryType":"{{queryType}}"}}
	`

	data := map[string]string{
		"merchantId": "141218",
		"queryType":  "businessInfo",
	}

	curlProxy := httpraw.HTTPProxy{
		HttpTpl: httpTpl,
		BeforRequest: func(reqDTO *httpraw.RequestDTO) (newReqDTO *httpraw.RequestDTO, err error) {
			reqDTO.Body = `{"_head":{"_interface":"NewMerchantCenterServer.Api.V1.getMerchantInfo","_msgType":"request","_remark":"","_version":"0.01","_timestamps":"1439261904","_invokeId":"563447634257324435","_callerServiceId":"210015","_groupNo":"1"},"_param":{"merchantId":"141218","queryType":"businessInfo"}}`
			return reqDTO, nil
		},
		AfterRequest: func(respDTO *httpraw.ResponseDTO) (newRespDTO *httpraw.ResponseDTO, err error) {
			respDTO.Body = gjson.Get(respDTO.Body, "_data").String()
			return respDTO, nil
		},
		LogInfoFn: func(reqDTO *httpraw.RequestDTO, respDTO *httpraw.ResponseDTO) {
			fmt.Println(reqDTO)
			fmt.Println(respDTO)
		},
	}
	ctx := context.Background()
	body, err := curlProxy.Proxy(ctx, data)
	require.NoError(t, err)
	fmt.Println(body)
}
