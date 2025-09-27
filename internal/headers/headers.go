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

var allowedSpecialChars = map[byte]struct{}{
	'!':  {},
	'#':  {},
	'$':  {},
	'%':  {},
	'&':  {},
	'\'': {},
	'*':  {},
	'+':  {},
	'-':  {},
	'.':  {},
	'^':  {},
	'_':  {},
	'`':  {},
	'|':  {},
	'~':  {},
}

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
	key = strings.ToLower(key)
	value := strings.TrimSpace(parts[1])

	if !validTokens([]byte(key)) {
		return 0, false, fmt.Errorf("invalid header token found: '%s'", key)
	}

	h.Set(key, value)
	bytesConsumed := idx + len(crlf)

	return bytesConsumed, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, exists := h[key]
	if exists {
		h[key] = fmt.Sprintf("%s, %s", v, value)
	} else {
		h[key] = value
	}
}

func validTokens(data []byte) bool {
	for _, c := range data {
		if !isTokenChar(c) {
			return false
		}
	}
	return true
}

func isTokenChar(c byte) bool {
	if c >= '0' && c <= '9' ||
		c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' {
		return true
	}
	_, exists := allowedSpecialChars[c]
	return exists
}
