package message

import (
	"fmt"
	"testing"
)

func TestEncodeMessage(t *testing.T) {
	m := Message{
		uuid(),
		Ping,
		uuid(),
		uuid(),
		[]byte("hello world!"),
	}

	b := EncodeMessage(m)
	fmt.Println(m)
	fmt.Println(b)
	m = DecodeMessage(b)
	fmt.Printf("content is %s, id is %x\n", string(m.Content), m.Identifier)
}
