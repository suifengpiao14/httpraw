package httpraw

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/funcs"
	"moul.io/http2curl"
)

func init() {
	mustache.AllowMissingVariables = false // 不允许变量缺失，变量缺失，不报错，会得到非法expr 表达式，将错误延迟到后续报错，增加使用、调试难度
}

const (
	Window_EOF           = "\r\n"
	Linux_EOF            = "\n"
	HTTP_HEAD_BODY_DELIM = Window_EOF + Window_EOF
)

type HttpTpl string

const (
	Http_header_Content_Type = "Content-Length"
)

// RenderTpl 解析模板，生成http raw 协议文本
func (htPt HttpTpl) RenderTpl(context ...any) (renderHttpRaw string, err error) {
	tpl := string(htPt)
	formatedTpl, err := FomrmatHttpRaw(tpl)
	if err != nil {
		return "", err
	}
	template, err := mustache.ParseStringRaw(formatedTpl, true)
	if err != nil {
		return "", err
	}
	context = append(context, TplRenderContext)
	rawHttp, err := template.Render(context...)
	if err != nil {
		return "", err
	}
	renderHttpRaw, err = FomrmatHttpRaw(rawHttp)
	if err != nil {
		return "", err
	}
	return renderHttpRaw, nil
}

// Request 解析模板，生成 http.Request 协议文本
func (htPt HttpTpl) Request(context ...any) (r *http.Request, err error) {
	rawHttp, err := htPt.RenderTpl(context...)
	if err != nil {
		return nil, err
	}
	r, err = ReadRequest(rawHttp)
	if err != nil {
		return nil, err
	}
	return r, nil
}
func (htPt HttpTpl) RequestTDO(context ...any) (reqDTO *RequestDTO, err error) {
	r, err := htPt.Request(context...)
	if err != nil {
		return nil, err
	}
	reqDTO, err = DestructReqeust(r)
	if err != nil {
		return nil, err
	}
	return reqDTO, nil
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

type RequestDTO struct {
	MetaData map[string]any `json:"metaData"` // metaData 用于存放一些额外的信息，例如请求的发起时间、循环次数、耗时等
	URL      string         `json:"url"`
	Method   string         `json:"method"`
	Header   http.Header    `json:"header"`
	Cookies  []*http.Cookie `json:"cookies"`
	Body     string         `json:"body"`
}

func (dto RequestDTO) String() string {
	b, err := json.Marshal(dto)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (dto RequestDTO) CurlCommand() string {
	req, err := dto.Request()
	if err != nil {
		return err.Error()
	}
	commd, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return err.Error()
	}
	curlCmd := commd.String()
	return curlCmd
}

func (rDTO RequestDTO) Copy() *RequestDTO {
	c := rDTO
	c.Header = copyHttpHeader(rDTO.Header)
	c.Cookies = make([]*http.Cookie, len(rDTO.Cookies))
	copy(c.Cookies, rDTO.Cookies)
	return &c
}

func (rDTO RequestDTO) GetCurlCmd() (curlCmd string, err error) {
	req, err := rDTO.Request()
	if err != nil {
		return "", err
	}
	curlCommand, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	curlCmd = curlCommand.String()
	return curlCmd, nil
}

func (rDTO RequestDTO) Request() (req *http.Request, err error) {
	req, err = BuildRequest(&rDTO)
	if err != nil {
		return nil, err
	}
	return req, nil
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
	MetaData   map[string]any `json:"metaData"` // metaData 用于存放一些额外的信息，例如请求的发起时间、循环次数、耗时等
	HttpStatus string         `json:"httpStatus"`
	Header     http.Header    `json:"header"`
	Cookies    []*http.Cookie `json:"cookies"`
	Body       string         `json:"body"`
	RequestDTO *RequestDTO    `json:"requestDTO"`
}

func (dto ResponseDTO) String() string {
	b, err := json.Marshal(dto)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (rDTO ResponseDTO) Copy() *ResponseDTO {
	c := rDTO
	c.Cookies = make([]*http.Cookie, len(rDTO.Cookies))
	copy(c.Cookies, rDTO.Cookies)
	c.RequestDTO = c.RequestDTO.Copy()
	c.Header = copyHttpHeader(rDTO.Header)

	return &c
}

func copyHttpHeader(header http.Header) (newHeader http.Header) {
	newHeader = make(http.Header)
	for k, v := range header {
		for _, v2 := range v {
			newHeader.Add(k, v2)
		}
	}
	return newHeader
}

func ParseResponse(HttpResponse []byte, r *http.Request) (responseDTO *ResponseDTO, err error) {
	byteReader := bytes.NewReader(HttpResponse)
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
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	responseDTO = &ResponseDTO{
		HttpStatus: strconv.Itoa(rsp.StatusCode),
		Header:     rsp.Header,
		Cookies:    rsp.Cookies(),
		Body:       string(body),
		RequestDTO: reqData,
	}
	return responseDTO, nil
}
