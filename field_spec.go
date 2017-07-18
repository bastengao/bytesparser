package bytesparser

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

const (
	// AttrDelimiter 表示属性分隔符
	AttrDelimiter = ","
	// keyValueDelimiter 表示键值对分隔符
	keyValueDelimiter = ":"
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
}

// 解析字段
func parseTag(tag string, instance interface{}) (FieldSpec, error) {
	fieldSpec := FieldSpec{}
	fieldSpec.Attrs = make(map[string]Attr)
	specs := strings.Split(tag, AttrDelimiter)
	for _, attr := range specs {
		strs := strings.Split(attr, keyValueDelimiter)
		k, v := strs[0], strs[1]
		switch k {
		case Len:
			if v == "*" {
				fieldSpec.Attrs[Len] = Attr{
					Name:         Len,
					IsExplicit:   true,
					IsExpression: false,
					Len:          -1,
				}
			} else {
				l, err := strconv.Atoi(v)
				if err == nil {
					fieldSpec.Attrs[Len] = Attr{
						Name:         Len,
						IsExplicit:   true,
						IsExpression: false,
						Len:          l,
					}
				} else {
					// try expression
					fieldSpec.Attrs[Len] = Attr{
						Name:         Len,
						IsExplicit:   true,
						IsExpression: true,
						Value:        v,
					}
				}
			}
		case Equal:
			if strings.Index(v, "0x") != 0 {
				return fieldSpec, errors.New("equal only support hex format")
			}

			src := []byte(v[2:])

			dst := make([]byte, hex.DecodedLen(len(src)))
			_, err := hex.Decode(dst, src)
			if err != nil {
				log.Fatal(err)
			}
			fieldSpec.Attrs[Equal] = Attr{
				Name:       Equal,
				IsExplicit: true,
				Equal:      dst,
			}
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
			fieldSpec.Attrs[Escape] = attr
		default:
			fieldSpec.Attrs[k] = Attr{
				Name:       k,
				IsExplicit: true,
				Value:      v,
			}
		}
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
			i++
		}
	}

	return escaped
}

func (s FieldSpec) canMatch() bool {
	// 如果 len 是通配符, 则不能直接匹配，需要其他字段匹配完了才能确定
	if s.Attrs[Len].Len == -1 {
		return false
	}

	return true
}

func (s FieldSpec) isMatch(c Context, offset int) (int, error) {
	// TODO: esacpe before

	// match len
	l, err := s.getLen(c, offset)
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

func (s FieldSpec) getLen(c Context, offset int) (int, error) {
	lenAttr, ok := s.Attrs[Len]
	if ok && lenAttr.Len != -1 {
		return lenAttr.getLen(c)
	}

	// TODO: auto detect len
	// reflect.TypeOf(a).Size()

	return 0, errors.New("could not ensure field length")
}

func (s FieldSpec) String() string {
	return fmt.Sprintf("{name: %s, offset: [%d:%d], len: %d}", s.Name, s.Start, s.End, s.End-s.Start)
}
