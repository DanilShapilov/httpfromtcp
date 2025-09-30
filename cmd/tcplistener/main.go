package main

import (
	"fmt"
	"log"
	"net"

	"github.com/DanilShapilov/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s", err)
	}
	defer l.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		// Wait on connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		fmt.Println("Accepted connection from", conn.RemoteAddr())

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error parsing request: %s\n", err.Error())
		}
		fmt.Println("Request line:")
		fmt.Println("- Method:", request.RequestLine.Method)
		fmt.Println("- Target:", request.RequestLine.RequestTarget)
		fmt.Println("- Version:", request.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
