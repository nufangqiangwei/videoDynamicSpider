package proxy

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type trojanServer struct {
	Password string `json:"password"`
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Sni      string `json:"sni"`
	Peer     string `json:"peer"`
	Ota      bool   `json:"ota"`
	Level    int    `json:"level"`
	Flow     string `json:"flow"`
	Method   string `json:"method"`
	Mark     string `json:"mark"`
}

var prefix = "trojan://"

const (
	at           = '@'
	colon        = ':'
	questionMark = '?'
	equal        = '='
	and          = '&'
	hashtag      = '#'
)

/*
trojan:// + password + @ + server + : + port + ? + sni + name
trojan://
    ba2ea9b1ce81@
    sg2.b96b642e58d2.airport.com:
    443?
    allowInsecure=1&
    peer=nhost.00cdn.com&
    sni=nhost.00cdn.com&
    type=tcp#%F0%9F%87%B8%F0%9F%87%AC%E6%96%B0%E5%8A%A0%E5%9D%A1%2002%20%7C%20%E4%B8%93%E7%BA%BF
*/
func trojanSerialization(trojanString string) (*trojanServer, error) {
	if !strings.HasPrefix(trojanString, prefix) {
		return nil, errors.New("trojan string must start with trojan://")
	}
	var (
		password       []byte
		server         []byte
		port           []byte
		argsKey        []byte
		argsValue      []byte
		args           map[string]string
		argsIdentifier byte = 0
		mark           []byte
		lastIdentifier byte = 0
		err            error
	)
	args = make(map[string]string)

	for index, i := range []byte(trojanString) {
		if index <= 8 {
			continue
		}
		if lastIdentifier == 0 {
			if i == at {
				lastIdentifier = at
				continue
			}
			password = append(password, i)
			continue
		}
		if lastIdentifier == at {
			if i == colon {
				lastIdentifier = colon
				continue
			}
			server = append(server, i)
			continue
		}
		if lastIdentifier == colon {
			if i == questionMark {
				lastIdentifier = questionMark
				continue
			}
			port = append(port, i)
			continue
		}
		if lastIdentifier == questionMark {
			if i == hashtag {
				lastIdentifier = hashtag
				continue
			}
			if i == equal {
				argsIdentifier = 1
				continue
			}
			if i == and {
				args[string(argsKey)] = string(argsValue)
				argsIdentifier = 0
				continue
			}
			if argsIdentifier == 0 {
				argsKey = append(argsKey, i)
				continue
			}
			if argsIdentifier == 1 {
				argsValue = append(argsValue, i)
				continue
			}
		}
		if lastIdentifier == hashtag {
			mark = append(mark, i)
		}
	}
	result := new(trojanServer)
	result.Password = string(password)
	result.Address = string(server)
	result.Port, err = strconv.Atoi(string(port))
	if err != nil {
		return nil, err
	}
	result.Sni, _ = args["sni"]
	result.Peer, _ = args["peer"]
	result.Flow, _ = args["flow"]
	result.Method, _ = args["method"]
	result.Mark, _ = url.QueryUnescape(string(mark))

	return result, nil
}

func ParseSubscriptionContainerDocument(rawConfig []byte) ([]string, error) {
	bodyDecoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(rawConfig))
	decoded, err := io.ReadAll(bodyDecoder)
	if err != nil {
		return nil, errors.New("failed to decode base64url body base64")
	}

	scanner := bufio.NewScanner(bytes.NewReader(decoded))

	const maxCapacity int = 1024 * 256
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	result := make([]string, 0)
	for scanner.Scan() {
		result = append(result, string(scanner.Bytes()))
	}
	return result, nil
}

func downloadSubData(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New("http status code is not 200")
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type TrojanConfig struct {
	Tag      string `json:"tag"`
	Protocol string `json:"protocol"`
	Settings struct {
		Servers []struct {
			Address  string `json:"address"`
			Method   string `json:"method"`
			Ota      bool   `json:"ota"`
			Password string `json:"password"`
			Port     int    `json:"port"`
			Level    int    `json:"level"`
			Flow     string `json:"flow"`
		} `json:"servers"`
	} `json:"settings"`
	StreamSettings struct {
		Network     string `json:"network"`
		Security    string `json:"security"`
		TlsSettings struct {
			AllowInsecure bool   `json:"allowInsecure"`
			ServerName    string `json:"serverName"`
		} `json:"tlsSettings"`
	} `json:"streamSettings"`
	Mux struct {
		Enabled     bool `json:"enabled"`
		Concurrency int  `json:"concurrency"`
	} `json:"mux"`
}

func getSubscription(url, tag string) *TrojanConfig {
	data, err := downloadSubData(url)
	if err != nil {
		return nil
	}
	var config = TrojanConfig{
		Tag:      tag,
		Protocol: "trojan",
		StreamSettings: struct {
			Network     string `json:"network"`
			Security    string `json:"security"`
			TlsSettings struct {
				AllowInsecure bool   `json:"allowInsecure"`
				ServerName    string `json:"serverName"`
			} `json:"tlsSettings"`
		}{
			Network:  "tcp",
			Security: "tls",
			TlsSettings: struct {
				AllowInsecure bool   `json:"allowInsecure"`
				ServerName    string `json:"serverName"`
			}{
				AllowInsecure: false,
			},
		},
		Mux: struct {
			Enabled     bool `json:"enabled"`
			Concurrency int  `json:"concurrency"`
		}{
			Enabled:     false,
			Concurrency: -1,
		},
		Settings: struct {
			Servers []struct {
				Address  string `json:"address"`
				Method   string `json:"method"`
				Ota      bool   `json:"ota"`
				Password string `json:"password"`
				Port     int    `json:"port"`
				Level    int    `json:"level"`
				Flow     string `json:"flow"`
			} `json:"servers"`
		}{
			Servers: []struct {
				Address  string `json:"address"`
				Method   string `json:"method"`
				Ota      bool   `json:"ota"`
				Password string `json:"password"`
				Port     int    `json:"port"`
				Level    int    `json:"level"`
				Flow     string `json:"flow"`
			}{},
		},
	}
	textContent, err := ParseSubscriptionContainerDocument(data)
	if err != nil {
		return nil
	}
	for _, i := range textContent {
		a, err := trojanSerialization(i)
		if err != nil {
			println(err.Error())
			continue
		}
		config.Settings.Servers = append(config.Settings.Servers, struct {
			Address  string `json:"address"`
			Method   string `json:"method"`
			Ota      bool   `json:"ota"`
			Password string `json:"password"`
			Port     int    `json:"port"`
			Level    int    `json:"level"`
			Flow     string `json:"flow"`
		}{
			Address:  a.Address,
			Method:   "chacha20",
			Ota:      false,
			Password: a.Password,
			Port:     a.Port,
			Level:    1,
			Flow:     "",
		})

	}
	return &config
}
