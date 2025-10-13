package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DanilShapilov/httpfromtcp/internal/headers"
	"github.com/DanilShapilov/httpfromtcp/internal/request"
	"github.com/DanilShapilov/httpfromtcp/internal/response"
	"github.com/DanilShapilov/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/video" {
		videoHandler(w, req)
		return
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return
	}
	handler200(w, req)
}

func videoHandler(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	const filepath = "assets/vim.mp4"
	videoBytes, err := os.ReadFile(filepath)
	if err != nil {
		handler500(w, nil)
		return
	}
	h := response.GetDefaultHeaders(len(videoBytes))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(videoBytes)
}

func proxyHandler(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)

	res, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	h.Override("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buf := make([]byte, maxChunkSize)
	bufTotal := make([]byte, 0)
	totalBytesRead := 0
	for {
		numBytesRead, err := res.Body.Read(buf)
		fmt.Println("Read", numBytesRead, "bytes")
		if numBytesRead > 0 {
			_, err = w.WriteChunkedBody(buf[:numBytesRead])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			bufTotal = append(bufTotal, buf[:numBytesRead]...)
			totalBytesRead += numBytesRead
		}

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body: ", err)
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}

	fmt.Println("BUF TOTAL LEN", len(bufTotal), "bytes")
	bufHash := sha256.Sum256(bufTotal)
	fmt.Printf("%x", bufHash)
	trailers := headers.NewHeaders()
	trailers.Override("X-Content-SHA256", fmt.Sprintf("%x", bufHash))
	trailers.Override("X-Content-Length", fmt.Sprintf("%v", totalBytesRead))

	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}
