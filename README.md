# "SHENZHEN GO" (working title)

SHENZHEN GO (working title) is an **experimental** visual Go environment, 
inspired by programming puzzle games such as TIS-100 and SHENZHEN I/O.

SHENZHEN GO provides a UI for editing a "graph," where the nodes are 
goroutines and the arrows are channel reads and writes. It can also convert 
a graph into Go source code which can be compiled and run. 

## Dependencies

SHENZHEN GO requires:

*   [Go](https://golang.org/)
*   [Graphviz](http://graphviz.org/)
*   A web browser (e.g. [Chrome](https://www.google.com/chrome)).

## Installation

This assumes you have set your `$GOPATH` (a common choice is `$HOME/go`, but it's up to you).

To install, open a terminal and run:

    go get -u github.com/google/shenzhen-go

This should create the `shenzhen-go` binary in your `$GOPATH/bin` directory.
Run it:

    $GOPATH/bin/shenzhen-go

and a web browser should appear with SHENZHEN GO (if not, navigate to 
http://[::1]:8088/ manually).

Navigate to the `$GOPATH/src/github.com/google/shenzhen-go/example.szgo` 
file and play around - this demonstrates an example program (a prime number sieve).

## Notes

This is not an official Google product.

This is an experimental project - expect plenty of rough edges and bugs, and no support.