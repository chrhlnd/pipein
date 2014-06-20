package main

import (
	"flag"
	"fmt"
	"github.com/chrhlnd/pipein"
	"log"
	"os"
	"os/signal"
)

func main() {
	var pipe *string
	pipe = flag.String("file", "/tmp/echopipe", "Specify the pipe to read on")

	shutdown := make(chan bool)

	log.Printf("Connecting ... %v ... ", *pipe)

	dchan, errs := pipein.NewPipeIn().Connect(*pipe, shutdown)

	if dchan == nil {
		log.Printf("Failed: %v", <-errs)
		os.Exit(1)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	done := false
	for !done {
		select {
		case s := <-sig:
			log.Printf("Got %v shutting down", s)
			done = true
			shutdown <- true
			break
		case d := <-dchan:
			fmt.Printf("%v", string(d))
			break
		case e := <-errs:
			log.Printf("ERR: %v", e)
			break
		}
	}

	<-shutdown
	log.Printf("Done!")
}
