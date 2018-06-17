# Getting Started

Contents:

* [Home](index.md)
* [Getting Started with v1](getting-started-v1.md)
* [Getting Started with v0](getting-started-v0.md)
* [Roadmap](roadmap.md)

## About versions

There are two versions of Shenzhen Go: the prototype that's deprecated, and the new version that isn't ready yet.
This guide describes the old deprecated prototype (v0).

## Dependencies

Shenzhen Go prototype (v0) requires:

*   [Go 1.7+](https://golang.org/)
*   [Graphviz](http://graphviz.org/)
*   A web browser (e.g. [Chrome](https://www.google.com/chrome)).

## Installation

If you are using Go 1.7, you need to have set your `$GOPATH` (common choices are `$HOME` and 
`$HOME/go`, but it's up to you). 
[For Go 1.8, the default `$GOPATH` is `$HOME/go`](https://rakyll.org/default-gopath/) so it
is not necessary to set it (but you can change it to override the default if you want).

To install, open a terminal and run:

    go get -u github.com/google/shenzhen-go/v0/cmd/shenzhen-go

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