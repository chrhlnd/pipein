package pipein

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"golang.org/x/exp/inotify"
)

type PipeConnection struct{}

func NewPipeIn() FifoReader {
	return &PipeConnection{}
}

const GROW_SIZE = 1024 * 4

func (p *PipeConnection) Connect(addr string, output chan<- []byte, shutdown chan bool) <-chan error {
	mine := false

	var err error
	var fi os.FileInfo

	errors := make(chan error, 1)

	if fi, err = os.Stat(addr); os.IsNotExist(err) {
		if err = syscall.Mknod(addr, syscall.S_IFIFO|0666, 0); err != nil {
			errors <- fmt.Errorf("Failed creating %v err %v", addr, err)
			return errors
		}
		mine = true
	} else if (fi.Mode() & os.ModeNamedPipe) == 0 {
		errors <- fmt.Errorf("%v must be a pipe", addr)
		return errors
	}

	var watcher *inotify.Watcher

	if watcher, err = inotify.NewWatcher(); err != nil {
		errors <- err
		return errors
	}

	if err = watcher.AddWatch(addr, inotify.IN_OPEN); err != nil {
		errors <- err
		return errors
	}

	var input *os.File

	readData := func() {
		buffer := make([]byte, GROW_SIZE)
		current := buffer[0:]
		pos := 0
		n := 0

	BUFFER:
		for {
			n, err = input.Read(current)

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
	}

	go func() {
		if input, err = os.Open(addr); err != nil {
			errors <- err
		} else {
		WORK:
			for {
				select {
				case <-watcher.Event:
					readData()
				case <-shutdown:
					log.Printf("Shutting down pipe %v", addr)
					break WORK
				}
			}
		}

		if err = input.Close(); err != nil {
			errors <- err
		}

		if mine {
			err := os.Remove(addr)
			if err != nil {
				errors <- fmt.Errorf("Err on %v err %v", addr, err)
			}
		}

		watcher.Close()

		shutdown <- true
	}()

	return errors
}
