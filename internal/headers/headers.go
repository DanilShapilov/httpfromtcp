package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// the empty line
		// headers are done, consume the CRLF
		return 2, true, nil
	}
	headerLineText := string(data[:idx])
	parts := strings.SplitN(headerLineText, ":", 2)
	if len(parts) != 2 {
		return 0, false, fmt.Errorf("error: incorrect headers format '%s'", headerLineText)
	}
	key := parts[0]
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: '%s'", key)
	}
	key = strings.TrimSpace(key)

	value := strings.TrimSpace(parts[1])

	h.Set(key, value)
	bytesConsumed := idx + len(crlf)

	return bytesConsumed, false, nil
}

func (h Headers) Set(key, value string) {
	h[key] = value
}
