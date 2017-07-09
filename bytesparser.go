// Package bytesparser implements parsing byte buff to struct
package bytesparser

import (
	"errors"
	"fmt"
	"reflect"
)

// Parse bytes to struct and return parsedIndex, err
func Parse(buff []byte, p interface{}) (parsedIndex int, err error) {
	parsedIndex = -1
	if reflect.TypeOf(p).Kind() != reflect.Ptr {
		err = errors.New("must be a pointer")
		return
	}

	context, err := newContext(buff, p)
	if err != nil {
		return parsedIndex, nil
	}

	offset, err := context.parse()
	if err != nil {
		return offset, err
	}
	fmt.Println(offset)

	return
}
