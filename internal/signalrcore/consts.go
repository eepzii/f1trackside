package signalrcore

// Connection states for the SignalR Core client.
// Defined as raw uint32 to allow direct use with atomic.Uint32 without type casting.
const (
	StateNew uint32 = iota
	StateConnecting
	StateConnected
	StateClosing
	StateClosed
)
