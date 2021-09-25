GOOS=linux 
go build -o output/crawler ./crawler/main.go
go build -o output/get ./get/main.go

zip output/crawler.zip output/crawler
zip output/get.zip output/get
