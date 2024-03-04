package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

// EncryptToken 使用aes进行加密
func EncryptToken(token, key string) string {
	body := map[string]string{}
	body["token"] = token
	body["time"] = strconv.FormatInt(time.Now().Unix(), 10)
	ip, _ := externalIP()
	body["ip"] = ip.String()
	data := []byte{}
	err := json.Unmarshal(data, body)
	if err != nil {
		return ""
	}
	// 加密
	aesKey := []byte(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		panic(err)
	}
	// 创建一个GCM模式的AES加密器
	// GCM模式提供了认证和加密功能
	// 需要一个唯一的、非重复的nonce（用于加密）和额外的认证数据（用于认证）
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	// 加密数据
	ciphertext := aesgcm.Seal(nil, nonce, data, nil)

	fmt.Printf("加密后的数据（十六进制）：%s\n", hex.EncodeToString(ciphertext))

	// 解密数据
	decrypted, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("解密后的数据：%s\n", decrypted)
	return hex.EncodeToString(ciphertext)
}

// DecryptToken 使用aes进行解密
func DecryptToken(token, key string) map[string]string {
	// 解密
	aesKey := []byte(key)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		panic(err)
	}
	// 创建一个GCM模式的AES加密器
	// GCM模式提供了认证和加密功能
	// 需要一个唯一的、非重复的nonce（用于加密）和额外的认证数据（用于认证）
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	// 解密数据
	ciphertext, err := hex.DecodeString(token)
	if err != nil {
		panic(err)
	}
	decrypted, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("解密后的数据：%s\n", decrypted)
	body := map[string]string{}
	err = json.Unmarshal(decrypted, &body)
	if err != nil {
		return nil
	}
	_, ok := body["token"]
	if !ok {
		body["token"] = ""
	}
	_, ok = body["time"]
	if !ok {
		body["time"] = ""
	}
	return body

}

// 获取公网IP地址
func externalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
