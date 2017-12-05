# "SHENZHEN GO" (working title)

[![Build Status](https://travis-ci.org/google/shenzhen-go.svg?branch=master)](https://travis-ci.org/google/shenzhen-go) [![Doc Status](https://godoc.org/github.com/google/shenzhen-go?status.svg)](https://godoc.org/github.com/google/shenzhen-go) [![license](https://img.shields.io/github/license/google/shenzhen-go.svg?maxAge=2592000)](https://github.com/google/shenzhen-go/blob/master/LICENSE)

SHENZHEN GO (working title) is an **experimental** visual Go environment, 
inspired by programming puzzle games such as TIS-100 and SHENZHEN I/O.

SHENZHEN GO provides a UI for editing a "graph," where the nodes are 
goroutines and the arrows are channel reads and writes. (This is analogous
to multiple "microcontrollers" communicating electrically in a circuit.)
It can also convert a graph into pure Go source code, which can be compiled 
and run, or used as a library in a regular Go program.

[SHENZHEN GO was unveiled](https://www.youtube.com/watch?v=AB9AUAmMlDo) at 
the [linux.conf.au 2017 Open Source & Games Miniconf](https://linux.conf.au/schedule/presentation/8/).

Read more at https://google.github.io/shenzhen-go.

![Example Graph](example_graph2.png)

## Dependencies

SHENZHEN GO requires:

*   [Go 1.7+](https://golang.org/)
*   [Graphviz](http://graphviz.org/)
*   A web browser (e.g. [Chrome](https://www.google.com/chrome)).

## Installation

If you are using Go 1.7, you need to have set your `$GOPATH` (common choices are `$HOME` and 
`$HOME/go`, but it's up to you). 
[For Go 1.8, the default `$GOPATH` is `$HOME/go`](https://rakyll.org/default-gopath/) so it
is not necessary to set it (but you can change it to override the default if you want).

To install, open a terminal and run:

    go get -u github.com/google/shenzhen-go/cmd/shenzhen-go

This should create the `shenzhen-go` binary in your `$GOPATH/bin` directory.
Run it:

    $GOPATH/bin/shenzhen-go

and a web browser should appear with SHENZHEN GO (if not, navigate to 
http://localhost:8088/ manually). 

The file browser is limited to the directory `shenzhen-go` was started in.

Navigate to the `examples/primes.szgo` file and play around - this demonstrates 
an example prime number sieve program.

## Notes

This is not an official Google product.

This is an experimental project - expect plenty of rough edges and bugs, and 
no support.

For discussions, there is [a Slack channel](https://gophers.slack.com/messages/shenzhen-go).
