package response

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/DanilShapilov/httpfromtcp/internal/headers"
)

const crlf = "\r\n"

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	var b strings.Builder
	for key, value := range headers {
		fmt.Fprintf(&b, "%s: %s%s", key, value, crlf)
	}
	b.WriteString(crlf)
	_, err := io.WriteString(w, b.String())
	if err != nil {
		return err
	}
	return nil
}
