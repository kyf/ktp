package session

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/kyf/ktp/message"
)

const (
	BUF_SIZE = 2 << 17 //1024 * 256
)

type Server struct {
	addr     string
	listener net.Listener
	logger   *log.Logger
}

func NewServer(addr string, logger *log.Logger) *Server {
	return &Server{addr: addr, logger: logger}
}

func (s *Server) Run() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer s.listener.Close()
	s.logger.Printf("server has listen %s", s.addr)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Printf("server.listener Accept err :%v", err)
			continue
		}
		go handleConn(conn, s)
	}
}

func responseErrCode(conn net.Conn, err error) {
	conn.Write([]byte(fmt.Sprintf("%v", err)))
}

type OnlineMap struct {
	sync.Mutex
	clients map[string]Client
}

func (this *OnlineMap) Add(conn net.Conn, uid message.UID) {
	this.Lock()
	defer this.Unlock()
	this.clients[string(uid[:])] = Client{uid, conn}
}

func (this *OnlineMap) Remove(uid message.UID) {
	this.Lock()
	defer this.Unlock()
	delete(this.clients, string(uid[:]))
}

var (
	OnMap *OnlineMap = &OnlineMap{clients: make(map[string]Client, 0)}
)

func registerMap(conn net.Conn, uid message.UID) {
	OnMap.Add(conn, uid)
}

func handleConn(conn net.Conn, s *Server) {
	defer conn.Close()

	//auth
	//connect msgtype
	connMsg, err := getConnectMessage(conn)
	if err != nil {
		responseErrCode(conn, err)
		return
	}
	//return connAck
	responseCode(conn, message.ConnAck, connMsg.From, connMsg.To)

	//register client
	registerMap(conn, connMsg.From)

	//heartbeat receive send
	//go heartbeat(conn, s)
	//disconnect

	buf := make([]byte, BUF_SIZE)
	for {
		num, err := conn.Read(buf)
		if err != nil {
			s.logger.Printf("conn Read err:%v", err)
			goto Exit
		}

		m := message.DecodeMessage(buf[:num])

		s.logger.Printf("receive data %v [%s]", m.Mtype, m.Content)

		switch m.Mtype {
		case message.Pong:

		case message.Disconn:
			OnMap.Remove(m.From)
		default:
		}

		//s.logger.Printf("receive data is %v", buf[:num])
	}

Exit:
}

func getConnectMessage(conn net.Conn) (*message.Message, error) {
	buf := make([]byte, 1024)
	num, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	content := buf[:num]
	m := message.DecodeMessage(content)
	return &m, nil
}

func responseCode(conn net.Conn, mtype message.MessageType, from, to message.UID, content ...string) {
	msg := ""
	if len(content) > 0 {
		msg = content[0]
	}
	m := message.Message{message.UUID(), mtype, from, to, []byte(msg)}
	conn.Write(message.EncodeMessage(m))
}
