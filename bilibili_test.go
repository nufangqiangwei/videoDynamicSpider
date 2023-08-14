package main

import (
	"fmt"
	"testing"
)

func TestGetVideo(t *testing.T) {
	spider := makeBilibiliSpider()
	for _, info := range spider.getVideoList() {
		fmt.Printf("%+v\n", info)
	}
}
