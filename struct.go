// +build !windows

package pipein

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

type PipeConnection struct{}

func NewPipeIn() FifoReader {
	return &PipeConnection{}
}

const GROW_SIZE = 1024 * 4

func (p *PipeConnection) Connect(addr string, output chan<- []byte, shutdown chan bool) <-chan error {
	errors := make(chan error, 1)

	//log.Print("Listening: ", addr)

	chanIn := make(chan *os.File)

	var lock sync.Mutex
	closed := false

	go func() {
		var conn *os.File

	SD:
		for {
			select {
				case <-shutdown: {
					//log.Println("Saw shutdown")
					lock.Lock()
					closed = true
					lock.Unlock()
					conn.Close()
					break SD
				}
				case cn := <-chanIn:
					conn = cn
			}
		}

		shutdown <- true
	}()

	{
		conn, err := os.OpenFile(addr, os.O_RDWR, os.ModeNamedPipe)
		if err != nil {
			if os.IsNotExist(err) {
				err = syscall.Mkfifo(addr, 0666)
				if err != nil {
					errors <- err
					return errors
				}
			} else {
				errors <- err
				return errors
			}
		} else {
			conn.Close()
		}
	}


	go func() {
		lclose := false

		for {
			lock.Lock()
			lclose = closed
			lock.Unlock()
			if lclose {
				return
			}

			conn, err := os.OpenFile(addr, os.O_RDWR, os.ModeNamedPipe)
			if err != nil {
				if err != io.EOF {
					panic(err.Error())
					errors <- err
				}
				return
			}

			chanIn <- conn

			buffer := make([]byte, GROW_SIZE)
			current := buffer[0:]
			pos := 0
			n := 0

			BUFFER:
			for {
				n, err = conn.Read(current)

				if err != nil && err != io.EOF {
					errors <- fmt.Errorf("Failed to read %v err %v", addr, err)
					break BUFFER
				}

				if n < len(buffer) {
					break BUFFER
				}

				pos = len(buffer)
				buffer = append(buffer, make([]byte, GROW_SIZE)...)
				current = buffer[pos:]
			}

			output <- buffer[0 : pos+n]

			conn.Close()
		}
	}()

	return errors
}
