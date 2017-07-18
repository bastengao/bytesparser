# bytesparser

## Install

```
go get github.com/bastengao/bytesparser
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/bastengao/bytesparser"
)

type Packet struct {
	Head    [2]byte `byte:"len:2,equal:0x55AA"`
	Command uint8  `byte:""`
	Len     uint16 `byte:"len:2,endian:little"`
	Data    []byte `byte:"len:{{.Len}}"`
}

func main() {
	packet := Packet{}
	buff := []byte{ 0x55, 0xAA, 0x01, 0x02, 0x00, 0x03, 0x04 }
	offset, err := bytesparser.Parse(buff, &packet)
	if err != nil {
        fmt.Println(err)
	}
	fmt.Println(packet, offset)
}
```

### "byte" tag options

* `len` indicates length of bytes used to parsing

    can be a number such as `byte:"len:2"`

	can be a wildcard such as `byte:"len:*"`, that indicates the length is not definite, depends other fields

	can be a expression such as `byte:"len:{{ .AnotherField }}"`

	can be optional, could auto detected if field size is definite

* `endian` indicates bigendian or littleendian. bigendian is default when endian not presents
   
   can be `little` or `big` such as `byte:"endian:little"`

   support uint8, uint16, uint32, uint64, int8, int16, int32, int64 and int(as int32)

* `equal` indicates the field equal specific value

   such as `byte:"equal:0xABCD"`, only support hex format

* `escape` indicated the field need unescape to parse
   
   must be a method of struct, such as `byte:"len:*,escape:escapeMethod"`. And the return value must be `map[byte][]byte`, key of map is original, value is escaped.

   ```go
   type Example struct {
	   Data []byte `byte:"len:*,escape:escapeMethod"`
   }

   func (_ Example) escapeMethod() map[byte][]byte {
	   return map[byte][]byte {
		   0x7D: {0x7D, 0x01},
		   0x7E: {0x7D, 0x02},
	   }
   }
   ```

### Supported type

* uint8, uint16, uint32, uint64 
* int8, int16, int32, int64 and int
* slice []byte or []uint
* array [n]byte or [n]uint

## TODO

* nested struct
* encode struct to bytes