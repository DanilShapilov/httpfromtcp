package response

import (
	"fmt"
)

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func getStatusLine(statusCode StatusCode) []byte {
	var reasonPhrases = map[StatusCode]string{
		200: "OK",
		400: "Bad Request",
		500: "Internal Server Error",
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s%s", statusCode, reasonPhrases[statusCode], crlf))
}
