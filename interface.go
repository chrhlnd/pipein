package pipein

// Reader connection to a system FIFO
type FifoReader interface {
	// Connect to a file fifo, shutdown channel is bidirectional -> means stop <- stopped
	Connect(addr string, output chan<- []byte, shutdown chan bool) <-chan error
}
