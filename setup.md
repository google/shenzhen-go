# Development setup

All the steps needed to completely rebuild (i.e. you want to edit client/ or proto/)

TODO(josh): Check this

1.  Install Go
2.  Install C++ protobuf compiler
3.  go get -u -v github.com/golang/protobuf
3.  go get -u -v github.com/gopherjs/gopherjs
4.  go get -u -v github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs
5.  go get -u -v github.com/improbable-eng/grpc-web/go/grpcweb
6.  go get -u -v github.com/google/shenzhen-go/...
7.  cd $GOPATH/src/github.com/google/shenzhen-go
8.  ./fullbuild.sh