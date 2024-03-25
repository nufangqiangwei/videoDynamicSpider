package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"videoDynamicAcquisition/utils"
)

var (
	manage          proxyManage
	GetConfigMethod LoadConfig
)

type LoadConfig interface {
	getProxyInfo() []utils.ProxyInfo
}

// Info 代理节点的信息
type Info struct {
	ssl           bool // 是否是https请求
	ip            string
	port          int
	token         string
	version       int
	pathPrefix    string
	supportMethod []proxyMethod
	runTaskNumber int
	mutex         sync.Mutex
	repeat        bool
}

func (i *Info) GetIp() string {
	return i.ip
}
func (i *Info) Unlock() {
	i.mutex.Unlock()
}

func (i *Info) newOrUpdate(info utils.ProxyInfo) {
	if i.ip == "" {
		i.ip = info.IP
		i.port = info.Port
		i.ssl = info.SSL
		i.token = info.Token
	} else {
		if i.ip != info.IP {
			panic("不允许覆盖更新代理节点信息")
		}
		if i.port != info.Port {
			i.port = info.Port
		}
		if i.ssl != info.SSL {
			i.ssl = info.SSL
		}
		if i.token != info.Token {
			i.token = info.Token
		}
	}
}

func (i *Info) Request(method string, args map[string]interface{}, responseStruct any) error {
	if !i.repeat {
		return LossOfUseRights{}
	}
	i.repeat = true
	var methodObject proxyMethod
	for _, info := range i.supportMethod {
		if info.Path == method {
			methodObject = info
			break
		}
	}
	if methodObject.Path == "" {
		return UndefinedMethod{method: method}
	}
	switch methodObject.httpMethod {
	case get:
		return methodObject.get(i, args, responseStruct)
	case post:
		return methodObject.post(i, args, responseStruct)
	}
	return HttpMethodRequestUndefined{path: method, method: methodObject.httpMethod}
}
func (i *Info) httpPrefix() string {
	var port, pathPrefix string
	if i.port == 80 || i.port == 443 {
		port = ""
	} else {
		port = fmt.Sprintf(":%d", i.port)
	}
	if i.pathPrefix == "" {
		pathPrefix = fmt.Sprintf("/%s", i.pathPrefix)
	} else {
		pathPrefix = i.pathPrefix
	}
	if i.ssl {
		return fmt.Sprintf("https://%s%s%s", i.ip, port, pathPrefix)
	}
	return fmt.Sprintf("http://%s%s%s", i.ip, port, pathPrefix)
}
func (i *Info) waitThreeSeconds() {
	time.Sleep(time.Second * 3)
	if !i.repeat {
		i.repeat = false
		i.Unlock()
	}
}

// proxyMethod 对方代码支持的方法，调用这个发起方法
type proxyMethod struct {
	httpMethod string
	Path       string
	syncMethod bool // 是否是异步的代理任务
}

func (pm proxyMethod) get(proxyInfo *Info, args map[string]interface{}, responseStruct any) error {
	return nil
}
func (pm proxyMethod) post(proxyInfo *Info, args map[string]interface{}, responseStruct any) error {
	requestBody, err := json.Marshal(args)
	if err != nil {
		return err
	}
	// 发送POST请求给代理
	url := fmt.Sprintf("%s%s", proxyInfo.httpPrefix(), pm.Path)
	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer func() {
		if !pm.syncMethod {
			proxyInfo.mutex.Unlock()
		}
	}()
	return json.Unmarshal(data, responseStruct)
}

type proxyManage struct {
	proxy map[string]*Info
	mutex sync.Mutex
}

func LoadConfigInfo() {
	if GetConfigMethod == nil {
		panic("缺少获取代理配置信息接口")
	}
	for _, proxy := range GetConfigMethod.getProxyInfo() {
		info, ok := manage.proxy[proxy.IP]
		if !ok {
			info = &Info{}
		}
		info.newOrUpdate(proxy)
		manage.proxy[proxy.IP] = info
	}
}
func GetMethodAvailableProxy() *Info {
	manage.mutex.Lock()
	defer manage.mutex.Unlock()
	for _, proxyInfo := range manage.proxy {
		if proxyInfo.mutex.TryLock() {
			proxyInfo.repeat = false
			return proxyInfo
		}
	}
	return nil
}
