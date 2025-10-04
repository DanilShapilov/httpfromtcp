package response

import (
	"strconv"

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
