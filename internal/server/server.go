package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/DanilShapilov/httpfromtcp/internal/request"
	"github.com/DanilShapilov/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytees := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytees))
	response.WriteHeaders(w, headers)
	w.Write(messageBytees)
}

// Server is an HTTP 1.1 server
type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerErr := &HandlerError{
			StatusCode: response.StatusCodeInternalServerError,
			Message:    err.Error(),
		}
		handlerErr.Write(conn)
		return
	}

	var buf bytes.Buffer
	handlerErr := s.handler(&buf, req)
	if handlerErr != nil {
		handlerErr.Write(conn)
		return
	}
	response.WriteStatusLine(conn, response.StatusCodeSuccess)
	headers := response.GetDefaultHeaders(buf.Len())
	response.WriteHeaders(conn, headers)
	fmt.Fprint(conn, &buf)
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := &Server{
		listener: l,
		handler:  handler,
	}

	go srv.listen()
	return srv, nil
}
