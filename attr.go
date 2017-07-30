package bytesparser

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

const (
	Len         = "len"
	Equal       = "equal"
	Endian      = "endian"
	Escape      = "escape"
	Wildcard    = "*"
	WildCardLen = -1
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

func newAttr(attrStr string, instance interface{}) (Attr, error) {
	strs := strings.Split(attrStr, KeyValueDelimiter)
	if len(strs) <= 1 {
		return Attr{}, errors.New("Could not parse")
	}

	k, v := strs[0], strs[1]
	switch k {
	case Len:
		if v == Wildcard {
			return Attr{
				Name:         Len,
				IsExplicit:   true,
				IsExpression: false,
				Len:          WildCardLen,
			}, nil
		} else {
			l, err := strconv.Atoi(v)
			if err == nil {
				return Attr{
					Name:         Len,
					IsExplicit:   true,
					IsExpression: false,
					Len:          l,
				}, nil
			} else {
				// try expression
				return Attr{
					Name:         Len,
					IsExplicit:   true,
					IsExpression: true,
					Value:        v,
				}, nil
			}
		}
	case Equal:
		if strings.Index(v, "0x") != 0 {
			return Attr{}, errors.New("equal only support hex format")
		}

		src := []byte(v[2:])

		dst := make([]byte, hex.DecodedLen(len(src)))
		_, err := hex.Decode(dst, src)
		if err != nil {
			log.Fatal(err)
		}
		return Attr{
			Name:       Equal,
			IsExplicit: true,
			Equal:      dst,
		}, nil
	case Escape:
		attr := Attr{
			Name:       Escape,
			IsExplicit: true,
			Value:      v,
		}
		value, err := attr.getEscape(instance)
		if err != nil {
			log.Fatal(err)
		}
		attr.Escape = value
		return attr, nil
	default:
		return Attr{
			Name:       k,
			IsExplicit: true,
			Value:      v,
		}, nil
	}
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
