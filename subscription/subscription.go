package subscription

import (
	"io"
	"net/http"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/subscription/myContainers"
	"videoDynamicAcquisition/subscription/myContainers/serverObj"
)

type subscriptionInfo struct {
	Tag             string
	Url             string
	ExpireSeconds   int
	UseProxyRequest bool
}

type Config struct {
	Source         []subscriptionInfo
	ImportProxyTag map[string]string
}

var subscriptionConfig *Config

func UpdateSubscription() error {
	//"vmess", "vless", "ss", "ssr", "trojan", "trojan-go", "http-proxy",
	//	"https-proxy", "socks5", "http2", "juicity", "tuic"
	proxyGroups := map[string][]serverObj.ServerObj{
		"vmess":       make([]serverObj.ServerObj, 0),
		"vless":       make([]serverObj.ServerObj, 0),
		"ss":          make([]serverObj.ServerObj, 0),
		"ssr":         make([]serverObj.ServerObj, 0),
		"trojan":      make([]serverObj.ServerObj, 0),
		"trojan-go":   make([]serverObj.ServerObj, 0),
		"http-proxy":  make([]serverObj.ServerObj, 0),
		"https-proxy": make([]serverObj.ServerObj, 0),
		"socks5":      make([]serverObj.ServerObj, 0),
		"http2":       make([]serverObj.ServerObj, 0),
		"juicity":     make([]serverObj.ServerObj, 0),
		"tuic":        make([]serverObj.ServerObj, 0),
	}

	for _, info := range subscriptionConfig.Source {
		response := getSubscriptionInfo(info)
		if response == nil {
			continue
		}
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Warning.Printf("read response error: %s", err)
			continue
		}
		for _, proxyInfo := range myContainers.TryAllParsers(responseBody) {
			proxyGroups[proxyInfo.ProtoToShow()] = append(proxyGroups[proxyInfo.ProtoToShow()], proxyInfo)
		}
	}
	return nil
}

func getSubscriptionInfo(source subscriptionInfo) *http.Response {
	req, err := http.NewRequest("GET", source.Url, nil)
	if err != nil {
		return nil
	}

	response, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil
	}

	return response
}

func SetConfig(config *Config) {
	subscriptionConfig = config
}
