package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
)

// isRetryableHTTPError reports whether err is a homeserver request that failed
// because Synapse was unreachable or overloaded, as opposed to rejecting the request.
func isRetryableHTTPError(err error) bool {
	var httpErr mautrix.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	if httpErr.Response == nil {
		// Request set: the request was sent and failed on the wire. Request nil means
		// a local marshal/build error, which retrying cannot fix.
		return httpErr.Request != nil
	}
	switch httpErr.Response.StatusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusTooManyRequests:
		return true
	}
	return false
}

// isRetryableMatrixError also accepts bare network and deadline errors.
func isRetryableMatrixError(err error) bool {
	if err == nil {
		return false
	}
	if isRetryableHTTPError(err) {
		return true
	}
	var netErr net.Error
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &netErr)
}

// exitForRestart stops the process so the next start backfills the message from
// Discord. The mautrix client has already retried; once those are spent the
// message would otherwise be dropped for good.
func exitForRestart(log *zerolog.Logger, err error, msg string) {
	log.Error().Err(err).Msg(msg + ", exiting so the restart's missed-message backfill recovers it")
	os.Exit(1)
}
