# json_marshal

自定义JSON编码，通过Tag中添加prec信息，对浮点数进行精度控制

## 使用说明
1. 保留两位小数`json:"price" prec:"2"`
2. 保留两位小数且转为字符串`json:"price" prec:"2,string"`

## 测试示例
```go
type TestObj struct {
		Price  float32 `json:"price" prec:"2,string"`
		Length float64 `json:"length" prec:"4"`
		Ignore int     `json:"-"`
		NoTag  string
}
t1 := TestObj{Price: 1.2345, Length: 100.2932999}
t2 := TestObj{Price: 34.2345, Length: 200.29}
output, err := JsonMarshalIndent(t1)
fmt.Println(string(output))

output, err = JsonMarshalIndent([]TestObj{t1, t2})
fmt.Println(string(output))

output, err = JsonMarshalIndent(map[string]any{
		"objs": []TestObj{t1, t2},
		"t1":   t1,
		"t2":   t2,
		"t3": map[string]any{
        "name":  "hello",
			  "price": 123.456,
		},
})
fmt.Println(string(output))
```
