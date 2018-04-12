# hak hak hak
Set-Location .\client
.\generate.ps1
Set-Location ..\server\view
go generate
Set-Location ..\..
go install github.com/google/shenzhen-go/dev/cmd/shenzhen-go
