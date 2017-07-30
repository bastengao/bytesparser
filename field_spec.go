package bytesparser

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// AttrDelimiter 表示属性分隔符
	AttrDelimiter = ","
	// KeyValueDelimiter 表示键值对分隔符
	KeyValueDelimiter = ":"
)

// FieldSpec 字段参数
type FieldSpec struct {
	Field reflect.StructField
	Name  string

	// 属性
	Attrs map[string]Attr

	Start int
	End   int
	Bytes []byte

	Context Context
}

// 解析字段
func parseTag(tag string, field reflect.StructField, instance interface{}) (FieldSpec, error) {
	fieldSpec := FieldSpec{Name: field.Name, Field: field}
	fieldSpec.Attrs = make(map[string]Attr)
	specs := strings.Split(tag, AttrDelimiter)
	for _, attrStr := range specs {
		attr, err := newAttr(attrStr, instance)
		if err != nil {
			continue
		}
		fieldSpec.Attrs[attr.Name] = attr
	}

	if fieldSpec.isStruct() {
		value := reflect.ValueOf(instance).Elem().FieldByName(field.Name)
		c, err := newContext([]byte{}, value.Addr().Interface())
		if err != nil {
			return fieldSpec, err
		}
		fieldSpec.Context = *c
	}

	return fieldSpec, nil
}

func (s FieldSpec) isBigEndian() (bigEndian bool) {
	bigEndian = true
	attr, ok := s.Attrs[Endian]
	if ok {
		if attr.Value == "little" {
			bigEndian = false
		}
	}
	return
}

func (s FieldSpec) needEscape() (ok bool) {
	_, ok = s.Attrs[Escape]
	return
}

func (s FieldSpec) escapedBytes() []byte {
	var escaped []byte
	attr, ok := s.Attrs[Escape]
	if !ok {
		return escaped
	}

	escapes := attr.Escape
	for i := 0; i < len(s.Bytes); {
		matched := false
		for k, v := range escapes {
			j := i + len(v)
			current := s.Bytes[i:j]
			if reflect.DeepEqual(current, v) {
				escaped = append(escaped, k)
				i = j
				matched = true
				break
			}
		}

		if !matched {
			escaped = append(escaped, s.Bytes[i])
			i++
		}
	}

	return escaped
}

func (s FieldSpec) canMatch() bool {
	// 如果 len 是通配符, 则不能直接匹配，需要其他字段匹配完了才能确定
	if attr, ok := s.Attrs[Len]; ok && attr.Len == WildCardLen {
		return false
	}

	return true
}

func (s FieldSpec) lastCanMatch() bool {
	context := s.Context
	return context.Specs[len(context.Specs)-1].canMatch()
}

func (s *FieldSpec) match(c Context, offset int) (int, error) {
	if s.isStruct() {
		s.Context.Buff = c.Buff[offset:]
		newOffset, err := s.Context.match()
		if err != nil {
			return newOffset, err
		}
		s.Start = offset
		s.End = offset + newOffset
		s.Bytes = c.Buff[offset : offset+newOffset]
		s.Context.Buff = s.Bytes
		return offset + newOffset, nil
	} else {
		newOffset, err := s.basicMatch(c, offset)
		if err != nil {
			return newOffset, err
		}
		s.Start = offset
		s.End = newOffset
		s.Bytes = c.Buff[offset:newOffset]
		return newOffset, nil
	}
}

func (s FieldSpec) basicMatch(c Context, offset int) (int, error) {
	// match len
	l, err := s.getLen(c)
	if err != nil {
		return l, err
	}

	if len(c.Buff) < offset+l {
		return 0, NeedMoreBytesError{Name: s.Name, Len: l}
	}

	// match eqaul
	if equal, ok := s.Attrs[Equal]; ok {
		end := offset + l
		if reflect.DeepEqual(c.Buff[offset:end], equal.Equal) {
			return end, nil
		}

		return -1, NotEqualError{Name: s.Name, Equal: equal.Equal}
		// return -1, fmt.Errorf("[%d:%d] not equal %v", offset, end, equal.Equal)
	}

	return offset + l, nil
}

func (s FieldSpec) getLen(c Context) (int, error) {
	lenAttr, ok := s.Attrs[Len]
	if ok && lenAttr.Len != WildCardLen {
		return lenAttr.getLen(c)
	}

	if s.isStruct() {
		totalLen := 0
		for _, spec := range s.Context.Specs {
			l, err := spec.getLen(s.Context)
			if err != nil {
				return 0, err
			}
			totalLen += l
		}
		return totalLen, nil
	} else {
		l, err := s.fieldSize()
		if err == nil {
			return l, nil
		}
	}

	return 0, fmt.Errorf("%s could not ensure field length", s.Name)
}

// 动态设置值, 根据值类型
func (s FieldSpec) setValue(c Context) bool {
	buff := s.Bytes
	if s.needEscape() {
		buff = s.escapedBytes()
	}

	fieldType := s.Field.Type
	value := reflect.ValueOf(c.Instance).Elem().FieldByName(s.Name)

	switch fieldType.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bigEndian := s.isBigEndian()
		v := toUint64(fieldType.Kind(), bigEndian, buff)
		value.SetUint(v)
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bigEndian := s.isBigEndian()
		v := toInt64(fieldType.Kind(), bigEndian, buff)
		value.SetInt(v)
		return true
	case reflect.Array:
		if fieldType.Elem().Kind() == reflect.Uint8 {
			reflect.Copy(value, reflect.ValueOf(buff))
			return true
		}
		fmt.Printf("array %s not supported yet\n", fieldType.Elem())
	case reflect.Slice:
		if fieldType.Elem().Kind() == reflect.Uint8 {
			value.SetBytes(buff)
			return true
		}

		fmt.Printf("slice %s not supported yet\n", fieldType.Elem())
	case reflect.Struct:
		// set by s.Context.match()
		return true
	}

	return false
}

func (s FieldSpec) isStruct() bool {
	return s.Field.Type.Kind() == reflect.Struct
}

func (s FieldSpec) String() string {
	if s.isStruct() {
		return fmt.Sprintf("{name: %s, type: %s, offset: [%d:%d], len: %d, bytes: %v, specs: %v}", s.Name, s.Field.Type, s.Start, s.End, s.End-s.Start, s.Bytes, s.Context.Specs)
	} else {
		return fmt.Sprintf("{name: %s, type: %s, offset: [%d:%d], len: %d, bytes: %v}", s.Name, s.Field.Type, s.Start, s.End, s.End-s.Start, s.Bytes)
	}
}

func (s FieldSpec) fieldSize() (int, error) {
	switch s.Field.Type.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Int,
		reflect.Array:
		return int(s.Field.Type.Size()), nil
	}

	return 0, fmt.Errorf("%s could not detect field size", s.Field.Type)
}
