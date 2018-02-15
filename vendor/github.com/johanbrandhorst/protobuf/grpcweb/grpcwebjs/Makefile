all: checkout build

checkout:
	npm install grpc-web-client webpack

build:
	./node_modules/.bin/webpack

clean:
	rm -f grpc.inc.js
	rm -rf node_modules
