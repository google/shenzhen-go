# Development setup

All the steps needed to completely rebuild (i.e. you want to make meaningful changes to anything but the server logic).

1.  Install Go
2.  Ensure `$GOPATH/bin` (`$HOME/go/bin`, by default) is in your `$PATH`
3.  Install [`protoc` (protobuf compiler)](https://github.com/protocolbuffers/protobuf/releases) into your path.
4.  Install [`mage`](https://magefile.org)
5.  If on Linux and want to use `-tags webview`: `sudo apt install webkit2gtk-4.0`
6.  If you haven't already, `go get -u -v github.com/google/shenzhen-go`
7.  `cd $GOPATH/src/github.com/google/shenzhen-go`
8.  `mage goGetTools`
9.  `mage build`