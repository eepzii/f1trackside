package signalrcore

import "errors"

var (
	errBufferOverflow     = errors.New("buffer overflow")
	errChannelUnavailable = errors.New("channel unavailable")
	errHandshake          = errors.New("handshake failed")
	errInvalidConfig      = errors.New("invalid configuration")
	errInvalidInput       = errors.New("invalid input parameter")
	errNegotiation        = errors.New("negotiation failed")
	errServerClosed       = errors.New("closed by server")
)
