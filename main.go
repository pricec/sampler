package main

import(
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"

	"github.com/go-yaml/yaml"
)

func sigHandler(
	cancel context.CancelFunc,
	sigChan <- chan os.Signal,
	wantExit *bool,
) {
	<- sigChan
	*wantExit = true
	cancel()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wantExit    := false
	sigChan     := make(chan os.Signal)

	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go sigHandler(cancel, sigChan, &wantExit)

	for !wantExit {
		fmt.Println("Waiting for context cancel")
		<-ctx.Done()
		fmt.Println("Got context cancel")
	}
	fmt.Println("Exiting")
}
