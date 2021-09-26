$Env:GOOS = "linux"
$Env:CGO_ENABLED = "0"
$Env:GOARCH = "amd64"

go build -o output/crawler ./crawler/main.go
go build -o output/get ./get/main.go
go build -o output/sns ./sns/main.go

~\Go\Bin\build-lambda-zip.exe -output output/crawler.zip output/crawler
~\Go\Bin\build-lambda-zip.exe -output output/get.zip output/get
~\Go\Bin\build-lambda-zip.exe -output output/sns.zip output/sns