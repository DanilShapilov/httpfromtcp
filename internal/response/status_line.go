package response

import (
	"fmt"
	"io"
)

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func getStatusLine(statusCode StatusCode) string {
	var reasonPhrases = map[StatusCode]string{
		200: "OK",
		400: "Bad Request",
		500: "Internal Server Error",
	}
	return fmt.Sprintf("HTTP/1.1 %d %s%s", statusCode, reasonPhrases[statusCode], crlf)
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := fmt.Fprint(w, getStatusLine(statusCode))
	if err != nil {
		return err
	}
	return nil
}
