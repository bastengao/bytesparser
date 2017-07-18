package bytesparser

import (
	"bytes"
	"encoding/binary"
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
		fieldSpec, err := parseTag(byteTag, p)
		fieldSpec.Name = field.Name
		fieldSpec.Field = field
		if err != nil {
			return nil, err
		}
		fieldSpecs = append(fieldSpecs, &fieldSpec)
	}

	var context = Context{
		Type:     t,
		Specs:    fieldSpecs,
		Instance: p,
		Buff:     buff,
	}
	return &context, nil
}

// -1 not match
// 0 match partly
// > 0 match fully, n present how many bytes matched in buff
func (c *Context) parse() (int, error) {
	offset, err := c.match()
	if err != nil {
		return offset, err
	}

	return offset, nil
}

/*
测试是否匹配
可以通过前后可以匹配的字段，算出中间的字段。但是不能解析连续两次都是 !canMatch
| canMatch | !canMatch | canMatch |
*/
func (c Context) match() (int, error) {
	var offset = 0
	var prevCanMatch = true
	for specIndex, spec := range c.Specs {
		var prevSpec *FieldSpec
		if specIndex > 0 {
			prevSpec = c.Specs[specIndex-1]
			prevCanMatch = prevSpec.canMatch()
			// 不能连续两次都不能解析
			if !prevCanMatch && !spec.canMatch() {
				return -1, fmt.Errorf("could not parse fields %s and %s", prevSpec.Name, spec.Name)
			}
		}
		if spec.canMatch() {
			// 如果之前的 field 是可以确定的，则值匹配一次, 否则遍历匹配
			if prevCanMatch {
				newOffset, err := spec.isMatch(c, offset)
				if err != nil {
					return newOffset, err
				}
				spec.Start = offset
				spec.End = newOffset
				spec.Bytes = c.Buff[offset:newOffset]
				offset = newOffset
				ok := c.setValue(*spec)
				if !ok {
					fmt.Println("could not set value")
				}
			} else {
				matched := false
				for i := offset; i < len(c.Buff); i++ {
					newOffset, err := spec.isMatch(c, i)
					if err != nil {
						continue
					}
					matched = true
					spec.Start = i
					spec.End = newOffset
					spec.Bytes = c.Buff[offset:newOffset]
					ok := c.setValue(*spec)
					if !ok {
						fmt.Println("could not set value")
					}
					offset = newOffset
					// 确定上一次不能解析字段的结束位置
					if prevSpec != nil {
						prevSpec.End = spec.Start
						prevSpec.Bytes = c.Buff[prevSpec.Start:prevSpec.End]
						ok := c.setValue(*prevSpec)
						if !ok {
							fmt.Println("could not set value")
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

// 动态设置值, 根据值类型
func (c Context) setValue(spec FieldSpec) bool {
	buff := spec.Bytes
	if spec.needEscape() {
		buff = spec.escapedBytes()
	}

	fieldType := spec.Field.Type
	value := reflect.ValueOf(c.Instance).Elem().FieldByName(spec.Name)

	switch fieldType.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bigEndian := spec.isBigEndian()
		v := toUint64(fieldType.Kind(), bigEndian, buff)
		value.SetUint(v)
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bigEndian := spec.isBigEndian()
		v := toInt64(fieldType.Kind(), bigEndian, buff)
		value.SetInt(v)
		return true
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.Uint8 {
			value.SetBytes(buff)
			return true
		}

		fmt.Printf("slice %s not supported yet\n", fieldType.Elem())
	}

	return false
}

func toInt64(kind reflect.Kind, bigEndian bool, buff []byte) int64 {
	var byteOrder binary.ByteOrder
	byteOrder = binary.BigEndian
	if !bigEndian {
		byteOrder = binary.LittleEndian
	}
	switch kind {
	case reflect.Int8:
		var i int8
		err := binary.Read(bytes.NewBuffer(buff), byteOrder, &i)
		if err != nil {
			return 0
		}
		return int64(i)
	case reflect.Int16:
		var i int16
		err := binary.Read(bytes.NewBuffer(buff), byteOrder, &i)
		if err != nil {
			return 0
		}
		return int64(i)
	case reflect.Int32:
		var i int32
		err := binary.Read(bytes.NewBuffer(buff), byteOrder, &i)
		if err != nil {
			return 0
		}
		return int64(i)
	case reflect.Int64:
		var i int64
		err := binary.Read(bytes.NewBuffer(buff), byteOrder, &i)
		if err != nil {
			return 0
		}
		return int64(i)
	case reflect.Int:
		switch len(buff) {
		case 1:
			return toInt64(reflect.Int8, bigEndian, buff)
		case 2, 3:
			return toInt64(reflect.Int16, bigEndian, buff)
		case 4, 5, 6, 7:
			return toInt64(reflect.Int32, bigEndian, buff)
		case 8:
			return toInt64(reflect.Int64, bigEndian, buff)
		default:
			return toInt64(reflect.Int64, bigEndian, buff)
		}

	default:
		return 0
	}
}

func toUint64(kind reflect.Kind, bigEndian bool, buff []byte) uint64 {
	switch kind {
	case reflect.Uint8:
		return uint64(buff[0])
	case reflect.Uint16:
		if bigEndian {
			return uint64(binary.BigEndian.Uint16(buff))
		}
		return uint64(binary.LittleEndian.Uint16(buff))
	case reflect.Uint32:
		if bigEndian {
			return uint64(binary.BigEndian.Uint32(buff))
		}
		return uint64(binary.LittleEndian.Uint32(buff))
	case reflect.Uint64:
		if bigEndian {
			return binary.BigEndian.Uint64(buff)
		}
		return uint64(binary.LittleEndian.Uint64(buff))
	default:
		return 0
	}
}
