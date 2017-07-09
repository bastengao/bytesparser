package bytesparser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// 常量
const (
	NeedMoreBytes = 0
	NotMatch      = -1

	Len    = "len"
	Equal  = "equal"
	Endian = "endian"
)

// Context 解析上下文
type Context struct {
	Buff     []byte
	Type     reflect.Type
	Specs    []FieldSpec
	Instance interface{}
}

func newContext(buff []byte, p interface{}) (*Context, error) {
	// parse all fields
	var fieldSpecs []FieldSpec
	t := reflect.ValueOf(p).Elem().Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		byteTag, ok := field.Tag.Lookup("byte")
		// skip
		if !ok {
			continue
		}
		fieldSpec, err := parseTag(byteTag)
		fieldSpec.Name = field.Name
		fieldSpec.Field = field
		if err != nil {
			return nil, err
		}
		fieldSpecs = append(fieldSpecs, fieldSpec)
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
func (c Context) parse() (int, error) {
	var offset = 0
	for _, spec := range c.Specs {
		newOffset, err := c.evaluateByFieldSpec(offset, spec)
		if err != nil {
			return newOffset, err
		}

		offset = newOffset
	}

	return offset, nil
}

func (c Context) evaluateByFieldSpec(offset int, spec FieldSpec) (int, error) {
	hasSetValue := false
	for key, attr := range spec.Attrs {
		switch key {
		case Len:
			l, err := attr.getLen(c)
			if err != nil {
				return -1, err
			}
			if len(c.Buff) < offset+l {
				return 0, fmt.Errorf("%s not enough length", spec.Name)
			}
		case Equal:
			end := offset + spec.Attrs[Len].Len
			if reflect.DeepEqual(c.Buff[offset:end], attr.Equal) {
				if !hasSetValue {
					if ok := c.setValue(spec, c.Buff[offset:end]); ok {
						hasSetValue = true
					}
				}
			} else {
				return -1, fmt.Errorf("[%d:%d] not equal %v", offset, end, attr.Equal)
			}
		}

		if !hasSetValue {
			l, err := spec.Attrs[Len].getLen(c)
			if err != nil {
				return -1, err
			}
			end := offset + l

			buff := c.Buff[offset:end]
			if ok := c.setValue(spec, buff); ok {
				hasSetValue = true
			}
		}
	}

	// TODO: Len must be set
	offset = offset + spec.Attrs[Len].Len
	return offset, nil
}

// 动态设置值, 根据值类型
func (c Context) setValue(spec FieldSpec, buff []byte) bool {
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
