# Getting Started

Contents
* [Home](index.md)
* [Getting Started](getting-started.md)
* [Roadmap](roadmap.md)

## Dependencies

SHENZHEN GO requires:

*   [Go 1.7+](https://golang.org/)
*   [Graphviz](http://graphviz.org/)
*   A web browser (e.g. [Chrome](https://www.google.com/chrome)).

## Installation

This assumes you have set your `$GOPATH` (common choices are `$HOME` and 
`$HOME/go`, but it's up to you).

To install, open a terminal and run:

    go get -u github.com/google/shenzhen-go/cmd/shenzhen-go

This should automatically download all the needed Go packages,
and create the `shenzhen-go` binary in your `$GOPATH/bin` directory.
Run it:

    $GOPATH/bin/shenzhen-go

and a web browser should appear with SHENZHEN GO (if not, navigate to 
http://localhost:8088/ manually). 

The file browser is limited to the directory `shenzhen-go` was started in.

Navigate to the `examples/primes.szgo` file and play around - this demonstrates 
an example prime number sieve program. Click the "Run" link and a list of 
prime numbers should be generated.