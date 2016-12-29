# Roadmap

Here's a bunch of TODOs in no particular order.

* Complete "new parts": Node editor that is based on the Part (does not assume parts.Code)
* Consider a different name for the project (or not!).
* Ability to delete nodes and channels
* More Parts, less code?
    * Filter
    * Function
* Add a type editor. (Or not, maybe just require writing and importing regular Go for that?)
* Expose a monitoring interface ("status page") for long-running programs that displays the same graph (perhaps annotated / coloured with completion of each goroutine).
* Create a library-ised version of the interface for use by the monitoring interface.
* Add interactively moving the nodes around on the surface - this would probably mean abandoning pure Graphviz.
* Add testing nodes / node shadows (static input or output for testing)
* Add a debugger.
* Add a formatter, linter, "vet".
* Decide on good practices.
* Tell more people about it.
