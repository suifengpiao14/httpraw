package httpraw

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
)

const (
	Window_EOF           = "\r\n"
	Linux_EOF            = "\n"
	HTTP_HEAD_BODY_DELIM = Window_EOF + Window_EOF
)

type httpTpl struct {
	Tpl      string
	Template *template.Template
}

const (
	Http_header_Content_Type = "Content-Length"
)

// FomrmatHttpRaw 格式化http 协议模板，手写协议在空格控制方面往往不规范，提供此方法，一是供内部格式化检测，二是给外部提供格式化途径
func FomrmatHttpRaw(httpRaw string) (formatHttpRaw string, err error) {
	httpRaw = funcs.TrimSpaces(httpRaw)
	lineArr := strings.Split(httpRaw, Linux_EOF)
	formatLineArr := make([]string, 0)
	headerContentType := strings.ToLower(Http_header_Content_Type)
	for _, line := range lineArr {
		formatLine := strings.TrimSpace(line)                                 // 去除每行的空格、制表符\r 等符号
		if strings.Contains(strings.ToLower(formatLine), headerContentType) { // 长度每次重新计算
			continue
		}
		formatLineArr = append(formatLineArr, formatLine)

	}
	httpRaw = strings.Join(formatLineArr, Window_EOF)
	if httpRaw == "" {
		err = errors.Errorf("http raw is empty")
		return "", err
	}

	headerRaw := strings.TrimSpace(httpRaw) // 默认只有请求头
	bodyRaw := ""                           // 默认body为
	bodyIndex := strings.Index(headerRaw, HTTP_HEAD_BODY_DELIM)
	if bodyIndex > -1 {
		headerRaw, bodyRaw = strings.TrimSpace(headerRaw[:bodyIndex]), strings.TrimSpace(headerRaw[bodyIndex:])
		bodyLen := len(bodyRaw)
		headerRaw = fmt.Sprintf("%s%s%s: %d", headerRaw, Window_EOF, Http_header_Content_Type, bodyLen)
	}
	formatHttpRaw = fmt.Sprintf("%s%s%s", headerRaw, HTTP_HEAD_BODY_DELIM, bodyRaw)
	// 检测模板是否符合 http 协议
	req, err := readRequestNoFormat(formatHttpRaw)
	if err != nil {
		return "", err
	}
	// 生成统一符合http 协议规范的模板
	reqByte, err := httputil.DumpRequest(req, true)
	if err != nil {
		return "", err
	}
	formatHttpRaw = string(reqByte)
	return formatHttpRaw, nil
}

// NewHttpTpl 实例化模版请求
func NewHttpTpl(tpl string) (*httpTpl, error) {
	formatedTpl, err := FomrmatHttpRaw(tpl)
	if err != nil {
		return nil, err
	}
	htPt := &httpTpl{
		Tpl: formatedTpl,
	}

	t, err := template.New("").Funcs(sprig.FuncMap()).Funcs(TemplatefuncMap).Parse(htPt.Tpl)
	if err != nil {
		return nil, err
	}
	htPt.Template = t
	return htPt, nil
}

// Request 解析模板，生成http raw 协议文本
func (htPt *httpTpl) Parse(data any) (rawHttp string, err error) {
	var b bytes.Buffer
	err = htPt.Template.Execute(&b, data)
	if err != nil {
		return
	}
	rawHttp = b.String()
	return rawHttp, nil
}

// Request 解析模板，生成http raw 协议文本
func (htPt *httpTpl) Request(data any) (r *http.Request, err error) {
	rawHttp, err := htPt.Parse(data)
	if err != nil {
		return nil, err
	}
	wellHttpRaw, err := FomrmatHttpRaw(rawHttp)
	if err != nil {
		return nil, err
	}
	r, err = ReadRequest(wellHttpRaw)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ReadRequest http 文本协议格式转http.Request 对象,需要格式化文本协议，请先调用 FomrmatHttpRaw 函数
func ReadRequest(httpRaw string) (req *http.Request, err error) {
	httpRaw, err = FomrmatHttpRaw(httpRaw) // 此处实现转换，确保格式ok(此处格式化，方便兼容手写)
	if err != nil {
		return nil, err
	}
	return readRequestNoFormat(httpRaw)
}

// readRequest http 文本协议格式转http.Request 对象,需要格式化文本协议，请先调用 FomrmatHttpRaw 函数
func readRequestNoFormat(httpRaw string) (req *http.Request, err error) {
	buf := bufio.NewReader(strings.NewReader(httpRaw))
	req, err = http.ReadRequest(buf)
	if err != nil {
		return
	}
	if req.URL.Scheme == "" {
		queryPre := ""
		if req.URL.RawQuery != "" {
			queryPre = "?"
		}
		req.RequestURI = fmt.Sprintf("http://%s%s%s%s", req.Host, req.URL.Path, queryPre, req.URL.RawQuery)
	}

	return req, nil
}

type RequestDTO struct {
	URL     string         `json:"url"`
	Method  string         `json:"method"`
	Header  http.Header    `json:"header"`
	Cookies []*http.Cookie `json:"cookies"`
	Body    string         `json:"body"`
}

// DestructReqeust 将 http.Request 转换为 request 结构体，方便将http raw 转换为常见的构造http请求参数
func DestructReqeust(req *http.Request) (requestDTO *RequestDTO, err error) {
	requestDTO = &RequestDTO{}
	var bodyByte []byte
	if req.Body != nil {
		bodyByte, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyByte))
	}

	req.Header.Del("Content-Length")
	requestDTO = &RequestDTO{
		URL:     req.URL.String(),
		Method:  req.Method,
		Header:  req.Header,
		Cookies: req.Cookies(),
		Body:    string(bodyByte),
	}

	return requestDTO, nil
}

func BuildRequest(requestDTO *RequestDTO) (req *http.Request, err error) {
	req, err = http.NewRequest(requestDTO.Method, requestDTO.URL, bytes.NewReader([]byte(requestDTO.Body)))
	if err != nil {
		return nil, err
	}
	requestDTO.Header.Del(Http_header_Content_Type) // 删除最初的长度头，使用新计算值
	for name, value := range requestDTO.Header {    // 循环赋值,确保不会覆盖 http.NewRequest自动生成的头信息
		for _, v := range value {
			req.Header.Add(name, v)
		}
	}
	for _, cookie := range requestDTO.Cookies {
		req.AddCookie(cookie)
	}
	return req, nil
}

type ResponseDTO struct {
	HttpStatus  string         `json:"httpStatus"`
	Header      http.Header    `json:"header"`
	Cookies     []*http.Cookie `json:"cookies"`
	Body        string         `json:"body"`
	RequestData *RequestDTO    `json:"requestData"`
}

func ParseResponse(b []byte, r *http.Request) (responseDTO *ResponseDTO, err error) {
	byteReader := bytes.NewReader(b)
	reader := bufio.NewReader(byteReader)
	rsp, err := http.ReadResponse(reader, r)
	if err != nil {
		return nil, err
	}
	reqData := new(RequestDTO)
	if r != nil {
		reqData, err = DestructReqeust(r)
		if err != nil {
			return nil, err
		}
	}
	responseDTO = &ResponseDTO{
		HttpStatus:  strconv.Itoa(rsp.StatusCode),
		Header:      rsp.Header,
		Cookies:     rsp.Cookies(),
		RequestData: reqData,
	}
	return responseDTO, nil
}
