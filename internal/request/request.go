package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/DanilShapilov/httpfromtcp/internal/headers"
)

const crlf = "\r\n"
const bufferSize = 8

type ParserState int

const (
	requestStateInitialized ParserState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	ParserState    ParserState
	bodyLengthRead int
}

func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == requestStateDone {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}
	totalBytesParsed := 0
	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case requestStateInitialized:
		rLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}
		if n == 0 {
			// just need more data
			return 0, nil
		}
		r.RequestLine = *rLine
		r.ParserState = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLenStr, exists := r.Headers.Get("content-length")
		if !exists {
			// assume that if no content-length header is present, there is no body
			r.ParserState = requestStateDone
			return len(data), nil
		}

		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", err)
		}

		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)

		if r.bodyLengthRead > contentLen {
			return 0, fmt.Errorf("error: request body length > Content-Length header value")
		}
		if r.bodyLengthRead == contentLen {
			r.ParserState = requestStateDone
		}
		return len(data), nil
	}

	return 0, fmt.Errorf("error: unknown state")
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		ParserState: requestStateInitialized,
		Headers:     headers.NewHeaders(),
		Body:        make([]byte, 0),
	}

	for req.ParserState != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				_, error := req.parse(buf[:readToIndex]) // in case we didn't even started
				if error != nil {
					return nil, error
				}
				if req.ParserState != requestStateDone {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + len(crlf), nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")

	if len(parts) != 3 {
		return nil, fmt.Errorf("request-line should match following format: method SP request-target SP HTTP-version")
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}
