package pipein

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

const GROW_SZ = 2048

type PipeConnection struct{}

func NewPipeIn() Reader {
	return &PipeConnection{}
}

func (p *PipeConnection) Connect(addr string, shutdown chan bool) (<-chan []byte, <-chan error) {
	mine := false
	done := false

	errors := make(chan error)

	if fi, err := os.Stat(addr); os.IsNotExist(err) {
		if err = syscall.Mknod(addr, syscall.S_IFIFO|0666, 0); err != nil {
			go func() {
				errors <- fmt.Errorf("Failed creating %v err %v", addr, err)
			}()
			return nil, errors
		}
		mine = true
	} else if (fi.Mode() & os.ModeNamedPipe) == 0 {
		go func() {
			errors <- fmt.Errorf("%v must be a pipe", addr)
		}()
		return nil, errors
	}

	output := make(chan []byte)

	writeDone := func() {
		done = true

		f, err := os.OpenFile(addr, os.O_WRONLY, 0)
		if err != nil {
			errors <- fmt.Errorf("Failed writing done to %v err %v", addr, err)
			return
		}

		_, err = f.Write([]byte("fin!"))
		if err != nil {
			errors <- fmt.Errorf("Failed sending done message to %v err %v", addr, err)
			return
		}
		err = f.Close()
		if err != nil {
			errors <- fmt.Errorf("Failed to close after writing done @ %v err %v", addr, err)
			return
		}
	}

	pump := func() {
		var buffer []byte
		var pos int
		var current []byte
		var n int

		for !done {
			buffer = make([]byte, GROW_SZ)
			current = buffer[0:]
			pos = 0

			f, err := os.Open(addr)
			if err != nil {
				errors <- fmt.Errorf("Failed to access %v err %v", addr, err)
				continue
			}

			if done {
				break
			}

			for !done {
				n, err = f.Read(current)
				if err != nil && err != io.EOF {
					errors <- fmt.Errorf("Failed to read %v err %v", addr, err)
				}
				if n < len(buffer) {
					break
				}
				pos = len(buffer)
				buffer = append(buffer, make([]byte, GROW_SZ)...)
				current = buffer[pos:]
			}

			err = f.Close()
			if err != nil {
				errors <- fmt.Errorf("Pump loop failed closing %v err %v", addr, err)
			}

			if done {
				break
			}

			output <- buffer
		}
	}

	go func() {
		go pump()

		select {
		case <-shutdown:
			done = true
			writeDone()
		}

		if mine {
			err := os.Remove(addr)
			if err != nil {
				errors <- fmt.Errorf("Failed removing %v err %v", addr, err)
			}
		}
		shutdown <- true
	}()

	return output, errors
}
