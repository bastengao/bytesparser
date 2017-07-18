package bytesparser

import "fmt"

// NeedMoreBytesError presents the buff length not big engouth to parse
type NeedMoreBytesError struct {
	Name string
	Len  int
}

func (e NeedMoreBytesError) Error() string {
	return fmt.Sprintf("%s need more bytes %d", e.Name, e.Len)
}

// NotEqualError presents the field not equal specific bytes
type NotEqualError struct {
	Name  string
	Equal []byte
}

func (e NotEqualError) Error() string {
	return fmt.Sprintf("%s not equal to %v", e.Name, e.Equal)
}
