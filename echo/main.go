package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/chrhlnd/pipein"
)

func main() {
	shutdown := make(chan bool)
	output := make(chan []byte, 5)

	errs := pipein.NewPipeIn().Connect("/tmp/incomming", output, shutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

WORK:
	for {
		select {
		case d := <-output:
			fmt.Printf("%v", string(d))
		case e := <-errs:
			fmt.Printf("ERR: %v\n", e)
		case <-sig:
			signal.Notify(sig)
			shutdown <- true
		DRAIN:
			for {
				select {
				case <-sig:
				default:
					break DRAIN
				}
			}
			break WORK
		}
	}

	<-shutdown
	fmt.Printf("Shutdown clean\n")
}
