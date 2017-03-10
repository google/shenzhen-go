#!/bin/bash
set -e
gopherjs build
go run ../../scripts/embed.go -p view -v svgResources -o ../static-svg.go svg.js svg.js.map
exit 0