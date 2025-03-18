package httpraw

import "time"

var TplRenderContext = map[string]any{}

func init() {
	// 注册内置上下文 支持{{Now.Unix}} {{Now.DataTime}} {{Now.NumberTime}}
	TplRenderContext["Now"] = NowTime{}
}

type NowTime struct{}

func (h NowTime) Unix() int64 {
	now := time.Now()
	// 获取秒级时间戳
	seconds := now.Unix()
	return seconds
}

func (h NowTime) DataTime() string {
	return time.Now().Format(time.DateTime)
}
func (h NowTime) NumberTime() string {
	return time.Now().Format("20060102150405")
}
