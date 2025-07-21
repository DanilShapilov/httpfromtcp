package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	f, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("couldn't open %s: %s", inputFilePath, err)
	}

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("=====================================")
	for line := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", line)
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
