package bytesparser

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

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
