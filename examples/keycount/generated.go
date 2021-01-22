// The keycount command was automatically generated by Shenzhen Go.
package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

var _ = runtime.Compiler

func Count_words(input <-chan string, output chan<- string, result chan<- map[string]uint) {

	defer func() {
		if output != nil {
			close(output)
		}
		close(result)
	}()

	m := make(map[string]uint)
	for in := range input {
		m[in]++
		if output != nil {
			output <- in
		}
	}
	result <- m
}

func Get_words(words chan<- string) {

	fmt.Println("Enter a line of text:")
	s, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	for _, word := range strings.Fields(s) {
		words <- word
	}
	close(words)
}

func Print_summary(result <-chan map[string]uint) {

	fmt.Printf("Got results: %v\n", <-result)
}

func main() {

	results := make(chan map[string]uint, 0)
	words := make(chan string, 0)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		Count_words(words, nil, results)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Get_words(words)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		Print_summary(results)
		wg.Done()
	}()

	// Wait for the various goroutines to finish.
	wg.Wait()
}