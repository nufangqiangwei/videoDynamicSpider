package proxy

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

const defaultProxyAddress = "http://192.168.0.20:1080"

var proxyStatus = false

// CheckProxyStatus 检查代理status
func CheckProxyStatus() {
	if !proxyStatus {
		checkProxyStatus()
	}
}

// 对defaultProxyAddress这个代理检查是否存在监听的服务，没有的话标记不可用
func checkProxyStatus() {
	proxy, err := url.Parse(defaultProxyAddress)
	if err != nil {
		proxyStatus = false
		return
	}
	// 尝试进行tcp连接
	coon, err := net.DialTimeout("tcp", proxy.Host, 3*time.Second)
	if err != nil {
		proxyStatus = false
		return
	}
	coon.Close()
	proxyStatus = true
}

func GetClient(useProxy bool) *http.Client {
	if !useProxy || !proxyStatus {
		return http.DefaultClient
	}
	proxy, _ := url.Parse(defaultProxyAddress)
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	return &http.Client{
		Transport: transport,
	}
}

func Init() {
	CheckProxyStatus()
}
