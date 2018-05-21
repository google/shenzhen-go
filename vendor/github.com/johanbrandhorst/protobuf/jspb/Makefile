all: checkout build

checkout:
	npm install google-protobuf webpack-cli webpack

build:
	./node_modules/.bin/webpack-cli

clean:
	rm jspb.inc.js
	rm -rf node_modules
