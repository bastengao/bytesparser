# bytesparser

## Install

```
go get github.com/bastengao/bytesparser
```

## Usage

```go
package main

import "github.com/bastengao/bytesparser"

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

## TODO

* byte escape 