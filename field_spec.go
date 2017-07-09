package bytesparser

import (
	"encoding/hex"
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
)

// FieldSpec 字段参数
type FieldSpec struct {
	Field reflect.StructField
	Name  string

	// 属性
	Attrs map[string]Attr
}

// 解析字段
func parseTag(tag string) (FieldSpec, error) {
	fieldSpec := FieldSpec{}
	fieldSpec.Attrs = make(map[string]Attr)
	specs := strings.Split(tag, ",")
	for _, attr := range specs {
		strs := strings.Split(attr, ":")
		k, v := strs[0], strs[1]
		switch k {
		case Len:
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
