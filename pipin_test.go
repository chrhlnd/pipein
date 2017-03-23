package pipein

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestPipeIn(t *testing.T) {
	output := make(chan []byte, 5)

	term := make(chan bool)

	finish := make(chan struct{})

	var errs <-chan error

	/*
		go func() {
			dur := time.Second * 3
			<-time.After(dur)
			panic(fmt.Sprintf("its been %v seconds", dur))
		}()
	*/

	go func() {
	INC:
		for {
			select {
			case data := <-output:
				log.Printf("<- %v", string(data))
			case _, ok := <-finish:
				if !ok {
					break INC
				}
			case err, ok := <-errs:
				log.Printf("ERR %v", err)
				if !ok {
					t.Fail()
					break INC
				}
			}

		}
	}()

	fname := "/tmp/test-pipe"

	connector := NewPipeIn()

	errs = connector.Connect(fname, output, term)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			log.Printf("Starting worker %v", worker)

			var err error
			var out *os.File
			if out, err = os.OpenFile(fname, os.O_APPEND|os.O_WRONLY, 0); err != nil {
				log.Print(err)
				t.Fail()
				return
			}

			defer out.Close()

			msg := fmt.Sprintf("I'm worker %v\n", worker)

			if _, err = out.Write([]byte(msg)); err != nil {
				log.Print(err)
				t.Fail()
				return
			}

			log.Print("Done ", msg)

		}(i)
	}
	wg.Wait()

	wg.Add(1)
	go func() {
		dur := time.Second * 15
		log.Printf("Push into %v will wait for %v", fname, dur)
		<-time.After(dur)
		wg.Done()
	}()
	wg.Wait()

	log.Print("Requesting shutdown")
	term <- true
	log.Print("Waiting shutdown")
	<-term
	close(term)
	close(finish)
}
