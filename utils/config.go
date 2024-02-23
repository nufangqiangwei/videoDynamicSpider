package utils

type ProxyInfo struct {
	IP    string `json:"IP"`
	Token string `json:"Token"`
}

type Config struct {
	DB struct {
		HOST         string `json:"HOST"`
		Port         int    `json:"Port"`
		User         string `json:"User"`
		Password     string `json:"Password"`
		DatabaseName string `json:"DatabaseName"`
	} `json:"DB"`
	Proxy                   []ProxyInfo `json:"Proxy"`
	DataPath                string
	ProxyDataRootPath       string `json:"ProxyDataRootPath"`
	ProxyWebServerLocalPort int    `json:"ProxyWebServerLocalPort"`
	Token                   string `json:"Token"`
}

const WaitImportPrefix = "waitImportFile"

const ImportingPrefix = "importingFile"

const FinishImportPrefix = "finishImportFile"

const ErrorImportPrefix = "errorImportFile"
