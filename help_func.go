package httpraw

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/suifengpiao14/yaegijson"
)

// SliceAny2string 将 []struct{}, []map[string]any 转成 []map[string]string
func SliceAny2string(structSlice any) (newData []map[string]string, err error) {
	rv := reflect.Indirect(reflect.ValueOf(structSlice))
	if rv.Kind() != reflect.Slice {
		err := errors.Errorf("required []struct{}, []map[string]any, but got :%T", structSlice)
		return nil, err
	}

	if rv.Len() == 0 {
		return newData, nil
	}

	for i := 0; i < rv.Len(); i++ {
		originRow := rv.Index(i).Interface()
		row := make(map[string]string)
		v := reflect.Indirect(reflect.ValueOf(originRow))
		switch v.Kind() {
		case reflect.Struct:
			for j := 0; j < v.NumField(); j++ {
				field := v.Type().Field(j)
				key := field.Tag.Get("json")
				if key == "" {
					key = field.Name
				}
				val := cast.ToString(v.Field(j).Interface())
				row[key] = val
			}
		case reflect.Map:
			for k, v := range originRow.(map[string]any) {
				row[k] = cast.ToString(v)
			}
		default:
			err := errors.Errorf("required struct or map , but got :%T", originRow)
			return nil, err

		}
		newData = append(newData, row)
	}

	return newData, nil
}

var GetValuesFromJson = yaegijson.GetValuesFromJson
var SetValueToJson = yaegijson.SetValueToJson

// DecodeResponseForJsonApiProtocol 解析jsonapi协议的响应数据(封装函数具有固定的语义，便于阅读理解)
func DecodeResponseForJsonApiProtocol(response, businessCodePath, businessMessagePath, dataPath string) (businessCode string, businessMessage string, data string) {
	values := yaegijson.GetValuesFromJson(response, businessCodePath, businessMessagePath, dataPath)
	businessCode, businessMessage, data = values[0], values[1], values[2]
	return businessCode, businessMessage, data
}
