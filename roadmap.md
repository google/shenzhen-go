# Roadmap

Contents:

* [Home](index.md)
* [Getting Started](getting-started.md)
* [Roadmap](roadmap.md)

Here's a bunch of TODOs in no particular order.

* Type-checking channel writes & reads
* Checking for channel usage where channels are passed as arguments to functions
* Report errors to the user better (for instance, highlighting nodes with broken syntax in them)
* Actually really really **for reals** write some documentation, Josh.
    * Title text on part menu
* Write an installer? (Obtains necessary components on demand? Or bundles them? Or one of each?)
* Quick channel creator (embed channel UI into a dropdown, default name, `interface{}`, cap 0)
* Other "kinds" of channel (e.g. `io.Pipe`)
* Improve the style and implementation of the UI (Polymer?)
* Consider a different name for the project (or not!).
* More Parts, less code!
    * Data repeater (read data, then repeat infinitely / n times)
    * /dev/null sink (reads until close, does nothing)
    * Matching sink (reads, compares to statically-defined data with optional looping - potential basis for testing)
    * Expression (`x -> f(x)`), allows for grouping function output into a struct, or sending on multiple different channels...
    * Map lookup (`x -> m[x]`), options for lookup failure (don't send, send 0 value, send `struct{T, bool}`...)
    * Multiplex / demultiplex / broadcast / first / last / ... (converting between 1 channel and many channels)
    * Statistics
    * Access to cmdline flags
    * Standard error logger
    * Text file reader (done!), text file writer
    * JSON file reader, JSON file writer
    * Database query, database update, database insert, database upsert
    * HTTP client (easy?), HTTP server (harder?)
    * Image file in, image file out
    * Email send, email read (add this last - see Zawinski)
* Part grouping
* Part patterns (based upon grouping)
    * 1-in 1-out chaining (polymerisation.... `[ Q ]_n = Q -> Q -> Q -> ... -> Q`)
    * Other patterns?
* Code snippets able to be edited in the user's favourite $EDITOR.
    * Maybe need a SZ-GO specific config file for options. `~/.szgo-rc`
    * And a config editor.
* Parts and channels are created with reasonable default names (unique against all existing components).
* Anonymous goroutines and channels?
* Parts that are created with the necessary channels attached.
    * The ability to merge channels together (Should be drag and drop operation?)
* Add a type editor. 
* Expose a monitoring interface ("status page") for long-running programs that displays the same graph (perhaps annotated / coloured with completion of each goroutine).
* Smarter channels
    * Channels display their current queue length (easy)
    * Channels display how many items went through (less easy)
* Create a library-ised version of the interface for use by the monitoring interface.
* Add interactively moving the nodes around on the surface - this would probably mean abandoning pure Graphviz.
* Add testing nodes / node shadows (static input or output for testing)
* Add a debugger.
* Add a formatter, linter, "vet".
* Decide on good practices.
* Tell more people about it.
