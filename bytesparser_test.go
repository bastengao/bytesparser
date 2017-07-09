package bytesparser

import (
	"fmt"
	"testing"
)

type SimplePacket struct {
	Head    []byte `byte:"len:2,equal:0x55AA"`
	Command uint8  `byte:"len:1,endian:big"`
	Len     uint16 `byte:"len:2,endian:little"`
	Len2    uint32 `byte:"len:4"`
	Len3    uint64 `byte:"len:8"`
}

func TestParse(t *testing.T) {
	packet := SimplePacket{}
	fmt.Println(packet)

	offset, err := Parse([]byte{0x55, 0xAA, 0x01, 0x02, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08}, &packet)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(packet)

	if offset != 17 {
		t.Error("failed")
	}

	if packet.Len != 2 {
		t.Error("failed")
	}

	if packet.Len2 != 4 {
		t.Error("failed")
	}

	if packet.Len3 != 8 {
		t.Error("failed")
	}
}

func TestParseOffset(t *testing.T) {
	packet := SimplePacket{}
	offset, err := Parse([]byte{0x55, 0xAB}, &packet)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(offset)

	if offset != NotMatch {
		t.Error("failed")
	}
}

func TestParseOffset2(t *testing.T) {
	packet := SimplePacket{}
	offset, err := Parse([]byte{0x55, 0xAA, 0x01}, &packet)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(offset)

	if offset != NeedMoreBytes {
		t.Error("failed")
	}
}
