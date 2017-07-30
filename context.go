package bytesparser

import (
	"errors"
	"fmt"
	"reflect"
)

// 常量
const (
	NeedMoreBytes = 0
	NotMatch      = -1
)

// Context 解析上下文
type Context struct {
	Buff     []byte
	Type     reflect.Type
	Specs    []*FieldSpec
	Instance interface{}

	PrevCanMatch bool
}

func newContext(buff []byte, p interface{}) (*Context, error) {
	// parse all fields
	var fieldSpecs []*FieldSpec
	t := reflect.ValueOf(p).Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		byteTag, ok := field.Tag.Lookup("byte")
		// skip
		if !ok {
			continue
		}
		fieldSpec, err := parseTag(byteTag, field, p)
		if err != nil {
			return nil, err
		}
		fieldSpecs = append(fieldSpecs, &fieldSpec)
	}

	var context = Context{
		Type:         t,
		Specs:        fieldSpecs,
		Instance:     p,
		Buff:         buff,
		PrevCanMatch: true,
	}
	return &context, nil
}

/*
-1 not match
0 match partly
> 0 match fully, n present how many bytes matched in buff

测试是否匹配
可以通过前后可以匹配的字段，算出中间的字段。但是不能解析连续两次都是 !canMatch
| canMatch | !canMatch | canMatch |
*/
func (c Context) match() (int, error) {
	var offset = 0
	var prevCanMatch = c.PrevCanMatch
	for specIndex, spec := range c.Specs {
		var prevSpec *FieldSpec
		if specIndex > 0 {
			prevSpec = c.Specs[specIndex-1]
			if prevSpec.isStruct() {
				prevCanMatch = prevSpec.lastCanMatch()
			} else {
				prevCanMatch = prevSpec.canMatch()
			}
			// 不能连续两次都不能解析
			if !prevCanMatch && !spec.canMatch() {
				return -1, fmt.Errorf("could not parse fields %s and %s", prevSpec.Name, spec.Name)
			}
		}
		if spec.canMatch() {
			if spec.isStruct() {
				spec.Context.PrevCanMatch = prevCanMatch
			}

			// 如果之前的 field 是可以确定的，则值匹配一次, 否则遍历匹配
			if prevCanMatch || specIndex == 0 {
				newOffset, err := spec.match(c, offset)
				if err != nil {
					return newOffset, err
				}
				offset = newOffset
				ok := spec.setValue(c)
				if !ok {
					fmt.Printf("%s could not set value", spec.Name)
				}
			} else {
				matched := false
				for i := offset; i < len(c.Buff); i++ {
					newOffset, err := spec.match(c, i)
					if err != nil {
						continue
					}
					matched = true
					ok := spec.setValue(c)
					if !ok {
						fmt.Printf("%s could not set value", spec.Name)
					}
					offset = newOffset
					// 确定上一次不能解析字段的结束位置
					if prevSpec != nil {
						fmt.Println("prev spec")
						prevSpec.End = spec.Start
						prevSpec.Bytes = c.Buff[prevSpec.Start:prevSpec.End]
						fmt.Println(prevSpec)
						// TODO: prevSpec.canMatch()
						ok := prevSpec.setValue(c)
						if !ok {
							fmt.Printf("%s could not set value", spec.Name)
						}
					}
					break
				}
				if !matched {
					return 0, errors.New("not match")
				}
			}
		} else {
			// 先确定开始位置
			if prevSpec != nil {
				spec.Start = prevSpec.End
			} else {
				spec.Start = offset
			}
			continue
		}
	}
	return offset, nil
}
