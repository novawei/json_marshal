package util

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

func indirect(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v
		} else {
			return v.Elem()
		}
	}
	return v
}

func isDestructableValue(v reflect.Value) bool {
	switch v.Type().String() {
	case "time.Time":
		return false
	default:
		k := v.Kind()
		return k == reflect.Struct || k == reflect.Slice || k == reflect.Array || k == reflect.Map
	}
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

func destruct(obj interface{}) (interface{}, error) {
	v := indirect(reflect.ValueOf(obj))
	switch v.Kind() {
	case reflect.Struct:
		output := map[string]interface{}{}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := indirect(v.Field(i))
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" || !value.CanInterface() {
				continue
			}
			if strings.Contains(jsonTag, "omitempty") && isEmptyValue(value) {
				continue
			}
			outputKey := strings.Split(jsonTag, ",")[0]
			if len(outputKey) == 0 {
				outputKey = field.Name
			}
			if isDestructableValue(value) {
				outputVal, err := destruct(value.Interface())
				if err != nil {
					return nil, err
				}
				output[outputKey] = outputVal
			} else {
				if value.Kind() == reflect.Float32 || value.Kind() == reflect.Float64 {
					precTag := field.Tag.Get("prec")
					if len(precTag) == 0 {
						output[outputKey] = value.Interface()
					} else {
						precStr := strings.Split(precTag, ",")[0]
						prec, err := strconv.ParseInt(precStr, 10, 32)
						if err != nil {
							return nil, err
						}
						f := value.Float()
						fstr := strconv.FormatFloat(f, 'f', int(prec), 64)
						if strings.Contains(precTag, "string") {
							output[outputKey] = fstr
						} else {
							output[outputKey], _ = strconv.ParseFloat(fstr, 64)
						}
					}
				} else {
					output[outputKey] = value.Interface()
				}
			}
		}
		return output, nil
	case reflect.Slice, reflect.Array:
		output := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			value := indirect(v.Index(i))
			if isDestructableValue(value) {
				outputVal, err := destruct(value.Interface())
				if err != nil {
					return nil, err
				}
				output = append(output, outputVal)
			} else {
				output = append(output, value.Interface())
			}
		}
		return output, nil
	case reflect.Map:
		output := map[string]interface{}{}
		for _, k := range v.MapKeys() {
			value := indirect(v.MapIndex(k))
			if isDestructableValue(value) {
				outputVal, err := destruct(value.Interface())
				if err != nil {
					return nil, err
				}
				output[k.String()] = outputVal
			} else {
				output[k.String()] = value.Interface()
			}
		}
		return output, nil
	default:
		return obj, nil
	}
}

// 自定义JSON编码，通过Tag中添加prec信息，对浮点数进行精度控制
//
// 使用说明:
//
//	1.保留两位小数`json:"price" prec:"2"`
//	2.保留两位小数且转为字符串`json:"price" prec:"2,string"`
//
// 测试示例:
//
//	type TestObj struct {
//		Price  float32 `json:"price" prec:"2,string"`
//		Length float64 `json:"length" prec:"4"`
//		Ignore int     `json:"-"`
//		NoTag  string
//	}
//	t1 := TestObj{Price: 1.2345, Length: 100.2932999}
//	t2 := TestObj{Price: 34.2345, Length: 200.29}
//	output, err := JsonMarshalIndent(t1)
//	fmt.Println(err)
//	fmt.Println(string(output))
//
//	output, err = JsonMarshalIndent([]TestObj{t1, t2})
//	fmt.Println(err)
//	fmt.Println(string(output))
//
//	output, err = JsonMarshalIndent(map[string]any{
//		"objs": []TestObj{t1, t2},
//		"t1":   t1,
//		"t2":   t2,
//		"t3": map[string]any{
//			"name":  "hello",
//			"price": 123.456,
//		},
//	})
//
//	fmt.Println(err)
//	fmt.Println(string(output))
func JsonMarshalIndent(obj interface{}) ([]byte, error) {
	val, err := destruct(obj)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(val, "", "    ")
}
