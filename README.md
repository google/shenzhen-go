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

## Versions

There are currently TWO versions of Shenzhen Go. 

The original Graphviz-based prototype is "v0", and mostly works, but I'm not working on it anymore. 

The second version, "v1", is not quite ready yet, so for now it lives in the "dev" directory. But I'm working on it.

## Getting started

See the getting-started guides at https://google.github.io/shenzhen-go.

## Notes

This is not an official Google product.

This is an experimental project - expect plenty of rough edges and bugs, and 
no support.

For discussions, there is [a Google Group](https://groups.google.com/forum/#!forum/szgo) and [a Slack channel](https://gophers.slack.com/messages/shenzhen-go).

## Acknowledgements

### v1 / dev

The dev version wouldn't be nearly as good as it is without the following:

* The [Ace](https://ace.c9.io/) code editor.
* [Chrome Hterm](https://chromium.googlesource.com/apps/libapps/+/master/hterm).
* [GopherJS](https://github.com/gopherjs/gopherjs).
* [gRPC](https://grpc.io/).
* [Improbable's gRPC-Web for Go](https://github.com/improbable-eng/grpc-web).
* Johan Brandhorst's [GopherJS bindings for gRPC-Web](https://github.com/johanbrandhorst/protobuf).

### v0

The prototype (v0) relies heavily on Graphviz to function, but does not bundle or vendor any Graphviz components.
