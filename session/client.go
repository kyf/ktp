package session

import (
	"fmt"
	"net"
	"time"

	"github.com/kyf/ktp/message"
)

type Client struct {
	uid  message.UID
	conn net.Conn
}

func NewClient(target string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", target, time.Second*30)
	if err != nil {
		return nil, err
	}

	uid := message.UUID()

	return &Client{uid, conn}, nil
}

func (c *Client) Reader(readChannel chan<- message.Message) {
	buf := make([]byte, BUF_SIZE)
	for {
		num, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		body := buf[:num]
		msg := message.DecodeMessage(body)
		readChannel <- msg
	}
}

func (c *Client) Connect() {
	m := message.Message{message.UUID(), message.Connect, c.uid, message.EmptyUUID(), nil}
	c.conn.Write(message.EncodeMessage(m))
}

func (c *Client) Send(content string) {
	m := message.Message{message.UUID(), message.Push, c.uid, message.UUID(), []byte(content)}
	c.conn.Write(message.EncodeMessage(m))
}

func (c *Client) Disconnect() {
	m := message.Message{message.UUID(), message.Disconn, c.uid, message.UUID(), nil}
	c.conn.Write(message.EncodeMessage(m))
	c.conn.Close()
}
