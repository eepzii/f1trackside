package signalrcore

import "errors"

var (
	errHandshake   = errors.New("handshake failed")
	errNegotiation = errors.New("negotiation failed")
)
