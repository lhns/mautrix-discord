package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"

	"maunium.net/go/mautrix"
)

func TestIsRetryableMatrixError(t *testing.T) {
	req := &http.Request{}
	resp := func(code int) *http.Response { return &http.Response{StatusCode: code} }
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"network error", mautrix.HTTPError{Request: req, WrappedError: errors.New("connection refused")}, true},
		{"wrapped network error", fmt.Errorf("x: %w", mautrix.HTTPError{Request: req}), true},
		{"marshal error", mautrix.HTTPError{Message: "failed to marshal JSON", WrappedError: errors.New("x")}, false},
		{"502", mautrix.HTTPError{Request: req, Response: resp(502)}, true},
		{"503", mautrix.HTTPError{Request: req, Response: resp(503)}, true},
		{"504", mautrix.HTTPError{Request: req, Response: resp(504)}, true},
		{"429", mautrix.HTTPError{Request: req, Response: resp(429)}, true},
		{"403", mautrix.HTTPError{Request: req, Response: resp(403)}, false},
		{"400", mautrix.HTTPError{Request: req, Response: resp(400)}, false},
		{"500", mautrix.HTTPError{Request: req, Response: resp(500)}, false},
		{"deadline", fmt.Errorf("x: %w", context.DeadlineExceeded), true},
		{"net error", &net.DNSError{IsTimeout: true}, true},
		{"plain error", errors.New("conversion failed"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableMatrixError(tt.err); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
	if isRetryableHTTPError(context.DeadlineExceeded) {
		t.Error("bare deadline error must not count as a homeserver HTTP error")
	}
}
