package httptemplate

import (
	"strconv"
	"strings"
	"text/template"
	"time"
)

var TemplatefuncMap = template.FuncMap{
	"zeroTime":      ZeroTime,
	"currentTime":   CurrentTime,
	"permanentTime": PermanentTime,
	"Contains":      strings.Contains,
	"fen2yuan":      Fen2yuan,
}

func ZeroTime() string {
	return "0000-00-00 00:00:00"
}

func CurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func PermanentTime() string {
	return "3000-12-31 23:59:59"
}

func Fen2yuan(fen interface{}) string {
	var yuan float64
	intFen, ok := fen.(int)
	if ok {
		yuan = float64(intFen) / 100
		return strconv.FormatFloat(yuan, 'f', 2, 64)
	}
	strFen, ok := fen.(string)
	if ok {
		intFen, err := strconv.Atoi(strFen)
		if err == nil {
			yuan = float64(intFen) / 100
			return strconv.FormatFloat(yuan, 'f', 2, 64)
		}
	}
	return strFen
}