// The http_server command was automatically generated by Shenzhen Go.
package main

import (
	"context"
	"github.com/google/shenzhen-go/dev/parts"
	"net/http"
	"sync"
)

func HTTPServer(addr <-chan string, errors chan<- error, requests chan<- *parts.HTTPRequest, shutdown <-chan context.Context) {
	const multiplicity = 1

	const instanceNumber = 0
	svr := &http.Server{
		Handler: parts.HTTPHandler(requests),
		Addr:    <-addr,
	}
	var shutdone chan struct{}
	go func() {
		ctx := <-shutdown
		shutdone = make(chan struct{})
		svr.Shutdown(ctx)
		close(shutdone)
	}()
	err := svr.ListenAndServe()
	if errors != nil {
		errors <- err
	}
	if shutdone != nil {
		<-shutdone
	}
	close(requests)
	if errors != nil {
		close(errors)
	}

}

func Hello_World(requests <-chan *parts.HTTPRequest) {
	const multiplicity = 1

	const instanceNumber = 0
	for rw := range requests {
		rw.Write([]byte("Hello, HTTP!\n"))
		rw.Close()
	}

}

func Send_8765(addr chan<- string) {
	const multiplicity = 1

	const instanceNumber = 0
	addr <- ":8765"

}

func main() {

	channel0 := make(chan *parts.HTTPRequest, 0)
	channel1 := make(chan string, 0)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		HTTPServer(channel1, nil, channel0, nil)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Hello_World(channel0)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Send_8765(channel1)
		wg.Done()
	}()

	// Wait for the end
	wg.Wait()
}
