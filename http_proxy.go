package httpraw

import (
	"context"
	"strconv"
)

type BeforeRequestFn func(reqDTO *RequestDTO) (newReqDTO *RequestDTO, err error)
type AfterRequestFn func(respDTO *ResponseDTO) (newRespDTO *ResponseDTO, err error)
type LogInfoFn func(reqDTO *RequestDTO, respDTO *ResponseDTO)

type HTTPProxy struct {
	HttpTpl         HttpTpl          `json:"httpTpl"`
	TransportConfig *TransportConfig `json:"transportConfig"`
	BeforRequest    BeforeRequestFn
	AfterRequest    AfterRequestFn
	LogInfoFn       LogInfoFn
}

func (proxy HTTPProxy) Proxy(ctx context.Context, context ...any) (responseBody string, err error) {
	r, err := proxy.HttpTpl.Request(context...)
	if err != nil {
		return "", err
	}
	reqDTO, err := DestructReqeust(r)
	if err != nil {
		return "", err
	}
	if proxy.BeforRequest != nil {
		reqDTO, err = proxy.BeforRequest(reqDTO)
		if err != nil {
			return "", err
		}
		r, err = reqDTO.Request()
		if err != nil {
			return "", err
		}
	}
	client := NewClient(proxy.TransportConfig)
	if proxy.LogInfoFn != nil {
		proxy.LogInfoFn(reqDTO.Copy(), nil)
	}
	responseBodyB, rsp, err := client.Execute(ctx, r)
	if err != nil {
		return "", err
	}
	responseBody = string(responseBodyB)
	responseDTO := &ResponseDTO{
		HttpStatus: strconv.Itoa(rsp.StatusCode),
		Header:     rsp.Header,
		Cookies:    rsp.Cookies(),
		Body:       responseBody,
		RequestDTO: reqDTO,
	}
	if proxy.AfterRequest != nil {
		responseDTO, err = proxy.AfterRequest(responseDTO)
		if err != nil {
			return "", err
		}
		responseBody = responseDTO.Body
	}
	if proxy.LogInfoFn != nil {
		proxy.LogInfoFn(nil, responseDTO.Copy())
	}
	return responseBody, nil
}
