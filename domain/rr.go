package domain

import "net/http"

// Request is an HTTP request.
type Request struct {
	Method  string
	Headers http.Header
	Path    string
	Body    []byte
}

// Response is an HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
}

// RRPair maps an HTTP request to an expected HTTP response and is
// critical for determining how successful an outbound request was.
type RRPair struct {
	Req  Request
	Resp Response
}
