package main

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

func checkNeedDestruct(k reflect.Kind) bool {
	return k == reflect.Struct || k == reflect.Slice || k == reflect.Array || k == reflect.Map
}

func unref(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func destruct(obj interface{}) (interface{}, error) {
	v := unref(reflect.ValueOf(obj))
	if v.Kind() == reflect.Struct {
		output := map[string]interface{}{}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := unref(v.Field(i))
			tag := field.Tag.Get("json")
			if len(tag) == 0 || strings.Contains(tag, "-") {
				continue
			}
			outputKey := strings.Split(tag, ",")[0]
			if checkNeedDestruct(value.Kind()) {
				outputVal, err := destruct(value.Interface())
				if err != nil {
					return nil, err
				}
				output[outputKey] = outputVal
			} else {
				if value.Kind() == reflect.Float32 || value.Kind() == reflect.Float64 {
					tag := field.Tag.Get("prec")
					if len(tag) == 0 {
						output[outputKey] = value.Interface()
					} else {
						precStr := strings.Split(tag, ",")[0]
						prec, err := strconv.ParseInt(precStr, 10, 32)
						if err != nil {
							return nil, err
						}
						f := value.Float()
						fstr := strconv.FormatFloat(f, 'f', int(prec), 64)
						if strings.Contains(tag, "string") {
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
	} else if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		output := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			value := unref(v.Index(i))
			if checkNeedDestruct(value.Kind()) {
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
	} else if v.Kind() == reflect.Map {
		output := map[string]interface{}{}
		for _, k := range v.MapKeys() {
			value := unref(v.MapIndex(k))
			if checkNeedDestruct(value.Kind()) {
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
	}
	return obj, nil
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

