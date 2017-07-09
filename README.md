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
	Head    []byte `byte:"len:2,equal:0x55AA"`
	Command uint8  `byte:"len:1"`
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

	can be a expression such as `byte:"len:{{ .AnotherField }}"`

* `endian` indicates bigendian or littleendian. bigendian is default when endian not presents
   
   can be `little` or `big` suchas `byte:"endian:little"`

   support uint8, uint16, uint32, uint64, int8, int16, int32, int64 and int(as int32)

* `equal` indicates the field equal specific value

   such as `byte:"equal:0xABCD"`, only support hex format

## TODO

* byte escape 
* nested struct
* auto detect field length
* encode struct to bytes