package httpraw_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suifengpiao14/httpraw"
)

func TestHttpTpl(t *testing.T) {
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

	rDTO, err := httpTpl.ReqeustTDO(data)
	require.NoError(t, err)
	fmt.Println(rDTO)

	req, err := rDTO.Request()
	require.NoError(t, err)
	//req1.Method, req1.URL.String(), req1.Body
	body, err := httpraw.RestyRequestFn(context.Background(), req, nil)
	require.NoError(t, err)

	fmt.Println(string(body))
}
