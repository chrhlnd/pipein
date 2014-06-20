pipein
=======

Small library to all for input on a fifo

Makes it easy to take input on pipe, and shut it down.

```
import (
	"github.com/chrhlnd/pipein"
	"os/signal"
	"os"
	"fmt"
)

func main() {
	shutdown := make(chan bool)
	
	data, errs := pipein.NewPipeIn().Connect("/tmp/incomming", shutdown)
	if data == nil {
		fmt.Printf("ERR: %v", <-errs)
		os.Exit(1)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	done := false

	for ; !done ; {
		select {
			case d := <-data:
				fmt.Printf("%v",string(d))
				break
			case <-sig:
				done = true
				shutdown <- true
				break
		}
	}

	<-shutdown
	fmt.Printf("Shutdown clean\n")
}
```

This is basiclly the echo program in sample. It captures data piped to the fifo. Then prints it.

On interrupt it cleanly shuts down.

The shutdown channel is bi directional.. Send to it to cause shutdown. Read from it to know its shutdown.


