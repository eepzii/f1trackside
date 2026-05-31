package signalrcore

import "errors"

var (
	errBufferOverflow     = errors.New("buffer overflow")
	errChannelUnavailable = errors.New("channel unavailable")
	errDuplicateListener  = errors.New("duplicate listener")
	errHandshake          = errors.New("handshake failed")
	errInvalidConfig      = errors.New("invalid configuration")
	errInvalidInput       = errors.New("invalid input parameter")
	errInvalidState       = errors.New("invalid client state")
	errInvocation         = errors.New("invocation failed")
	errNegotiation        = errors.New("negotiation failed")
	errNotConnected       = errors.New("client is not connected")
	errServerClosed       = errors.New("closed by server")
)
