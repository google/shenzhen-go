# SHENZHEN GO

Contents
* [Home](index.md)
* [Getting Started](getting-started.md)
* [Roadmap](roadmap.md)

SHENZHEN GO (working title) is an **experimental** visual Go environment, 
inspired by programming puzzle games such as TIS-100 and SHENZHEN I/O.

SHENZHEN GOs provides a UI for editing a "graph," where the nodes are 
goroutines and the arrows are channel reads and writes. This is analogous
to multiple "microcontrollers" communicating electrically in a circuit.
It can also convert a graph into pure Go source code, which can be compiled 
and run, or used as a library in a regular Go program.

[SHENZHEN GO was unveiled](https://www.youtube.com/watch?v=AB9AUAmMlDo) at 
the [linux.conf.au 2017 Open Source & Games Miniconf](https://linux.conf.au/schedule/presentation/8/).

![Example Graph](example_graph2.png)

See the [Getting Started](getting-started.md) guide to start using it.

## Notes

This is not an official Google product.

This is an experimental project - expect plenty of rough edges and bugs, and 
no support.

## More notes

*   SHENZHEN GO is (for now) a strictly one-way process. You *cannot* import Go code 
    that you wrote into SHENZHEN GO. 
*   You *can* write snippets of Go in your SHENZHEN GO graph, which then appear 
    in the Go output.
*   One day it should be possible to write zero Go code, yet produce wonderful 
    graphs that do useful things.
*   You can always save a copy of your program as Go, continue working on that, 
    and never touch SHENZHEN GO again. 
*   However, modifications to the generated output won't be preserved if 
    SHENZHEN GO builds or runs the design again.
*   Don't treat the Go output as a virtuous paragon of how to code in Go. It is
    "machine-generated" and therefore held to a lower standard than "hand-made".
*   The JSON-based file format aims to be *diffable*, or at least *not ugly*, 
    for the benefit of source control and code review.
