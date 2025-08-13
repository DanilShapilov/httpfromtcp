package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		for line := range getLinesChannel(conn) {
			fmt.Printf("%s\n", line)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)
	go func() {
		defer f.Close()
		defer close(c)
		s := ""
		bytes := make([]byte, 8)
		for {
			n, err := f.Read(bytes)
			if err != nil {
				if s != "" {
					c <- s
					s = ""
				}
				if errors.Is(err, io.EOF) {
					return
				}
				log.Fatalf("error reading file: %v", err)
				return
			}
			parts := strings.Split(string(bytes[:n]), "\n")
			for i := 0; i < len(parts)-1; i++ {
				line := fmt.Sprint(s, parts[i])
				c <- line
				s = ""
			}
			s += parts[len(parts)-1]
		}
	}()

	return c
}
