package session

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kyf/ktp/message"
)

const (
	BUF_SIZE = 2 << 17 //1024 * 256
)

var (
	PongWait time.Duration = time.Second * 10

	PingPeriod time.Duration = (PongWait * 8) / 10
)

type Server struct {
	addr     string
	listener net.Listener
	logger   *log.Logger
}

func NewServer(addr string, logger *log.Logger) *Server {
	return &Server{addr: addr, logger: logger}
}

func httponline() {
	http.HandleFunc("/online", func(w http.ResponseWriter, r *http.Request) {
		for _, client := range OnMap.clients {
			w.Write([]byte(fmt.Sprintf("%s<br>\n", client.uid)))
		}
	})
	log.Print(http.ListenAndServe(":1233", nil))
}

func (s *Server) Run() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go httponline()

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

func heartbeat(conn net.Conn, s *Server, from, to message.UID) {
	ticker := time.NewTicker(PingPeriod)
	for {
		select {
		case <-ticker.C:
			m := message.Message{message.UUID(), message.Ping, from, to, nil}
			conn.Write(message.EncodeMessage(m))
		}
	}
}

func handleConn(conn net.Conn, s *Server) {
	var connFrom *message.UID
	defer func() {
		conn.Close()
		if connFrom != nil {
			OnMap.Remove(*connFrom)
		}
	}()

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
	connFrom = &connMsg.From

	//heartbeat receive send
	go heartbeat(conn, s, connMsg.From, connMsg.To)

	conn.SetReadDeadline(time.Now().Add(PongWait))

	buf := make([]byte, BUF_SIZE)
	for {
		num, err := conn.Read(buf)
		if err != nil {
			s.logger.Printf("conn Read err:%v", err)
			goto Exit
		}

		m := message.DecodeMessage(buf[:num])

		s.logger.Printf("receive data %v [%s], from:[%v], to:[%v]", m.Mtype, m.Content, m.From, m.To)
		switch m.Mtype {
		case message.Pong:
			conn.SetReadDeadline(time.Now().Add(PongWait))
		case message.Disconn:
			OnMap.Remove(m.From)
		default:
			if client, ok := OnMap.clients[string(m.To[:])]; ok {
				m.Mtype = message.Receive
				client.conn.Write(message.EncodeMessage(m))
			}
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
