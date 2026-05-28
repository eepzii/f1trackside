package signalrcore

import "errors"

var (
	errBufferOverflow     = errors.New("buffer overflow")
	errChannelUnavailable = errors.New("channel unavailable")
	errHandshake          = errors.New("handshake failed")
	errNegotiation        = errors.New("negotiation failed")
	errServerClosed       = errors.New("closed by server")
)
