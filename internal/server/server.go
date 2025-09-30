package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

// Server is an HTTP 1.1 server
type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		// Wait on connection
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response := "HTTP/1.1 200 OK\r\n" + // Status line
		"Content-Type: text/plain\r\n" + // Example header
		"Content-Length: 13\r\n" + // Content length header
		"\r\n" + // Blank line to separate headers from the body
		"Hello World!\n" // Body
	conn.Write([]byte(response))
}

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := &Server{
		listener: l,
	}

	go srv.listen()
	return srv, nil
}
