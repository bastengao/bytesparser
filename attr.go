package bytesparser

import (
	"bytes"
	"strconv"
	"text/template"
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
