package proxy

import (
	"net/http"
	"net/url"
)

const defaultProxyAddress = "http://127.0.0.1:1080"

func GetClient(useProxy bool) *http.Client {
	if !useProxy {
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
