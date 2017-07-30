package bytesparser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCanMatch(t *testing.T) {
	field := reflect.StructField{Name: "Test", Type: reflect.TypeOf(1)}
	fieldSpec, err := parseTag("", field, nil)
	if err != nil {
		fmt.Println(nil)
	}

	if !fieldSpec.canMatch() {
		t.Error("failed")
	}
}

func TestGetLen(t *testing.T) {
	field := reflect.StructField{Name: "Test", Type: reflect.TypeOf(1)}
	fieldSpec, err := parseTag("", field, nil)

	if err != nil {
		fmt.Println(nil)
	}

	fmt.Println(fieldSpec)
}
