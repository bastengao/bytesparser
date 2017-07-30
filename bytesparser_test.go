package bytesparser

import (
	"fmt"
	"reflect"
	"testing"
)

type SimplePacket struct {
	Head    []byte `byte:"len:2,equal:0x55AA"`
	Command uint8  `byte:"len:1,endian:big"`
	Len     uint16 `byte:"len:2,endian:little"`
	Len2    uint32 `byte:"len:4"`
	Len3    uint64 `byte:"len:8"`
}

type DelimiterPacket struct {
	Head    uint8  `byte:"len:1,equal:0x7E"`
	Payload []byte `byte:"len:*,escape:Escapes"`
	Tail    uint8  `byte:"len:1,equal:0x7E"`
}

type BigPacket struct {
	Head       byte       `byte:""`
	NestPacket NestPacket `byte:""`
}

type NestPacket struct {
	Len uint16 `byte:""`
}

type BigDelimiterPacket struct {
	Head        byte        `byte:""`
	Data        []byte      `byte:"len:*"`
	NestePacket Nest2Packet `byte:""`
}

type Nest2Packet struct {
	Len uint16 `byte:"len:2,equal:0x0003"`
}

func (p DelimiterPacket) Escapes() map[byte][]byte {
	return map[byte][]byte{
		0x7D: {0x7D, 0x01},
		0x7E: {0x7D, 0x02},
	}
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

func TestParseWithDelimiter(t *testing.T) {
	packet := DelimiterPacket{}
	offset, err := Parse([]byte{0x7E, 0x7D, 0x02, 0x7D, 0x01, 0x7E}, &packet)
	if err != nil {
		fmt.Println(err)
		t.Error("failed")
	}

	if offset != 6 {
		t.Error("failed")
	}

	if !reflect.DeepEqual(packet.Payload, []byte{0x7E, 0x7D}) {
		t.Error("failed")
	}

	fmt.Println(packet)
}

func TestNestedStruct(t *testing.T) {
	packet := BigPacket{}
	offset, err := Parse([]byte{0x01, 0x00, 0x02}, &packet)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(packet, offset)
}

func TestNestedDelimiterStruct(t *testing.T) {
	packet := BigDelimiterPacket{}
	offset, err := Parse([]byte{0x01, 0x02, 0x00, 0x03}, &packet)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(packet, offset)
}
