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
	ProxyDataRootPath       string
	ProxyWebServerLocalPort int
	Token                   string
}

const WaitImportPrefix = "waitImportFile"

const ImportingPrefix = "importingFile"

const FinishImportPrefix = "finishImportFile"

const ErrorImportPrefix = "errorImportFile"
