# Roadmap

Here's a bunch of TODOs in no particular order.

* Consider a different name for the project (or not!).
* Add saving / loading of programs.
* Add / remove buttons for nodes and edges.
* Change the way channels are represented: they should be nodes as well, since many goroutines can access them (not merely one source and one destination).
* Extract the relationship between goroutines and channels by parsing the Go directly.
* Add a type editor. (Or not, maybe just require writing and importing regular Go for that?)
* Expose a monitoring interface ("status page") for long-running programs that displays the same graph (perhaps annotated / coloured with completion of each goroutine).
* Create a library-ised version of the interface for use by the monitoring interface.
* Add interactively moving the nodes around on the surface - this would probably mean abandoning GraphViz.
* Add a debugger.
* Add a formatter, linter, "vet".
* Decide on good practices.
* Tell more people about it.
