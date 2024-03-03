.PHONY: spider webServer

spider:
	go build -o ./program/videoInfoSpider ./cmd/spider
webServer:
	go build -o ./program/videoServer ./cmd/spiderProxy