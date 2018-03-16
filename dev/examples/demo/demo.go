package demo

import "fmt"

// Demo is a function that reads from foo and writes 42 to bar.
func Demo(foo <-chan int, bar chan<- int) {
	fmt.Printf("Got %d, sending 42\n", <-foo)
	bar <- 42
}
