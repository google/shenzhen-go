$env:GOOS="linux"
gopherjs build
$env:GOOS="windows"
go run ../../scripts/embed.go -p view -v clientResources -o ../server/view/static-client.go client.js*