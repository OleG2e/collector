package network

import (
	"bytes"
	"io"
	"net/http"
)

type RequestInfo struct {
	Method string
	URL    string
	Body   string
}

func NewRequestInfo(req *http.Request) *RequestInfo {
	var bodyBuffer bytes.Buffer
	req.Body = io.NopCloser(io.TeeReader(req.Body, &bodyBuffer))

	b, _ := io.ReadAll(req.Body)

	req.Body = io.NopCloser(&bodyBuffer)

	return &RequestInfo{
		Method: req.Method,
		URL:    req.URL.String(),
		Body:   string(b),
	}
}

func (i *RequestInfo) String() string {
	return i.Method + " " + i.URL + " " + i.Body
}
