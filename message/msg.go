package message

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

type MessageType byte

type UID [12]byte

func (uid UID) String() string {
	return fmt.Sprintf("%x", uid[:])
}

func UUID() UID {
	return uuid()
}

func EmptyUUID() UID {
	return UID([12]byte{})
}

type Message struct {
	Identifier UID
	Mtype      MessageType
	From       UID
	To         UID
	Content    []byte
}

func NewMessageString(m string) *Message {
	return &Message{
		uuid(),
		Connect,
		uuid(),
		uuid(),
		[]byte(m),
	}
}

func EncodeMessage(msg Message) []byte {
	result := bytes.NewBuffer(make([]byte, 0))
	binary.Write(result, binary.LittleEndian, msg.Identifier)
	binary.Write(result, binary.BigEndian, msg.Mtype)
	binary.Write(result, binary.BigEndian, msg.From)
	binary.Write(result, binary.BigEndian, msg.To)
	binary.Write(result, binary.BigEndian, msg.Content)

	return result.Bytes()
}

func DecodeMessage(data []byte) (msg Message) {
	begin, end := 0, 12

	r := bytes.NewReader(data[begin:end])
	binary.Read(r, binary.BigEndian, &msg.Identifier)
	begin = end
	end += 1

	r = bytes.NewReader(data[begin:end])
	binary.Read(r, binary.BigEndian, &msg.Mtype)
	begin = end
	end += 12

	r = bytes.NewReader(data[begin:end])
	binary.Read(r, binary.BigEndian, &msg.From)
	begin = end
	end += 12

	r = bytes.NewReader(data[begin:end])
	binary.Read(r, binary.BigEndian, &msg.To)
	begin = end

	r = bytes.NewReader(data[begin:])
	msg.Content = make([]byte, len(data[begin:]))
	binary.Read(r, binary.BigEndian, &msg.Content)
	return
}

const (
	Connect MessageType = iota
	ConnAck
	Ping
	Pong
	Push
	Receive
	Disconn
)

func (this MessageType) String() string {
	switch this {
	case Connect:
		return "Connect"
	case ConnAck:
		return "ConnAck"
	case Ping:
		return "Ping"
	case Pong:
		return "Pong"
	case Push:
		return "Push"
	case Receive:
		return "Receive"
	case Disconn:
		return "Disconn"
	}

	return ""
}

func machineId() []byte {
	var id []byte = make([]byte, 3)
	hostname, err := os.Hostname()
	if err != nil {
		io.ReadFull(rand.Reader, id)
		return id
	}

	re := md5.Sum([]byte(hostname))
	copy(id, re[:])
	return id
}

var MachineId = machineId()

var uuidCounter uint32 = 0

func uuid() UID {
	var b [12]byte
	//timestamp--4
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	//machine--3
	b[4] = MachineId[0]
	b[5] = MachineId[1]
	b[6] = MachineId[2]
	//pid--2
	pid := os.Getpid()
	b[7] = byte(pid >> 16)
	b[8] = byte(pid >> 8)
	//increment--3
	atomic.AddUint32(&uuidCounter, 1)
	b[9] = byte(uuidCounter >> 16)
	b[10] = byte(uuidCounter >> 8)
	b[11] = byte(uuidCounter)
	return UID(b)
}
