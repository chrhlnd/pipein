package pipein

// Reader connection to a system FIFO
type Reader interface {
	// Connect to a file fifo, shutdown channel is bidirectional -> means stop <- stopped
	Connect(addr string, shutdown chan bool) (<-chan []byte, <-chan error)
}
