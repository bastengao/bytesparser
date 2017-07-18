package bytesparser

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"text/template"
)

const (
	Len    = "len"
	Equal  = "equal"
	Endian = "endian"
	Escape = "escape"
)

// Attr 属性
type Attr struct {
	Name         string
	IsExplicit   bool
	IsExpression bool
	Value        interface{}

	Len int

	// 值等于, 目前只支持十六进制, 例如 0x55AA
	Equal []byte

	// 转义映射, key 是转义前的值，value  是转义后的值，目前只支持原始值是一个字节的情况
	Escape map[byte][]byte
}

func (attr Attr) getValue(c Context) (interface{}, error) {
	switch attr.Name {
	case Len:
		return attr.getLen(c)
	case Equal:
		return attr.Equal, nil
	}

	return nil, nil
}

func (attr Attr) getLen(c Context) (int, error) {
	if !attr.IsExpression {
		return attr.Len, nil
	}

	templ, err := template.New("attr").Parse(attr.Value.(string))
	if err != nil {
		return -1, err
	}

	result := new(bytes.Buffer)
	err = templ.Execute(result, c.Instance)
	i, err := strconv.Atoi(result.String())
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (attr Attr) getEscape(i interface{}) (map[byte][]byte, error) {
	if attr.Name != Escape {
		return *new(map[byte][]byte), errors.New("is not escape")
	}

	methodName := attr.Value.(string)

	meth, ok := reflect.TypeOf(i).Elem().MethodByName(methodName)
	if !ok {
		return *new(map[byte][]byte), errors.New("escape method not found")
	}

	ins := reflect.ValueOf(i).Elem()
	values := meth.Func.Call([]reflect.Value{ins})
	if len(values) < 1 {
		return map[byte][]byte{}, errors.New("not value returned")
	}

	return values[0].Interface().(map[byte][]byte), nil
}
