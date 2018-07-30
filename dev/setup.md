# Development setup

All the steps needed to completely rebuild (i.e. you want to edit `client/*` or `proto/*`).

1.  Install Go
2.  Ensure `$GOPATH/bin` (`$HOME/go/bin`, by default) is in your `$PATH`
3.  Install `protoc` (protobuf compiler)
4.  Install various commands:  
    1.  `go get -u -v github.com/golang/dep/cmd/dep`
    2.  `go get -u -v github.com/golang/protobuf/protoc-gen-go`
    3.  `go get -u -v github.com/gopherjs/gopherjs`
    4.  `go get -u -v github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs`
4.  (a) If on Linux: `sudo apt install webkit2gtk-4.0`
5.  If you haven't already, `go get -u -v github.com/google/shenzhen-go/dev/cmd/shenzhen-go`
6.  `cd $GOPATH/src/github.com/google/shenzhen-go/dev`
7.  `./fullbuild.sh`