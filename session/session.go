package session

import (
	"log"
	"net"

	"github.com/kyf/ktp/message"
)

const (
	BUF_SIZE = 1024 * 256
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

func handleConn(conn net.Conn, s *Server) {
	defer conn.Close()

	buf := make([]byte, BUF_SIZE)
	for {
		num, err := conn.Read(buf)
		if err != nil {
			s.logger.Printf("conn Read err:%v", err)
			goto Exit
		}

		m := message.DecodeMessage(buf[:num])

		s.logger.Printf("receive data %v [%s]", m.Mtype, m.Content)

		//s.logger.Printf("receive data is %v", buf[:num])
	}

Exit:
}
