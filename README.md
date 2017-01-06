# "SHENZHEN GO" (working title)

SHENZHEN GO (working title) is an **experimental** visual Go environment, inspired by programming puzzle games such as TIS-100 and SHENZHEN I/O.

It structures a Go program as a visual graph where the nodes are goroutines and the arrows are channels. 

## Dependencies

SHENZHEN GO requires:

*   A recent Go installation.
*   Graphviz installed.
*   A web browser.

## Installation

This assumes you have set your `$GOPATH` (a common choice is `$HOME/go`, but it's up to you).

To install, run:

    go get -u github.com/google/shenzhen-go

This should create the `shenzhen-go` binary in your `$GOPATH/bin` directory.
Run it and a web browser should appear with SHENZHEN GO (or navigate to 
http://[::1]:8088/ manually).

Navigate to the `$GOPATH/src/github.com/google/shenzhen-go/example.szgo` 
file and play around - this demonstrates an example program (a prime number sieve).

## Notes

This is not an official Google product.

This is an experimental project - expect plenty of rough edges and bugs, and no support.