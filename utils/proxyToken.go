package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
)

// EncryptToken 使用aes进行加密
func EncryptToken(token, key, ivStr string) string {
	ciphertext := cbcEncrypter([]byte(token), []byte(key), []byte(ivStr))
	ciphertextStr := hex.EncodeToString(ciphertext)
	return ciphertextStr
}

// DecryptToken 使用aes进行解密
func DecryptToken(token, key, ivStr string, body any) error {
	println("token：", token)
	ciphertext, err := hex.DecodeString(token)
	if err != nil {
		return err
	}
	unpaddedPlaintext := cbcDecrypter(ciphertext, []byte(key), []byte(ivStr))

	fmt.Println("Decrypted:", string(unpaddedPlaintext))
	return json.Unmarshal(unpaddedPlaintext, &body)

}

/*
	CBC 加密
	text 待加密的明文
    key 秘钥
*/
func cbcEncrypter(text []byte, key []byte, iv []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	// 填充
	paddText := pkcs7Padding(text, block.BlockSize())

	blockMode := cipher.NewCBCEncrypter(block, iv)

	// 加密
	result := make([]byte, len(paddText))
	blockMode.CryptBlocks(result, paddText)
	// 返回密文
	return result
}

/*
	CBC 解密
	encrypter 待解密的密文
	key 秘钥
*/
func cbcDecrypter(encrypter []byte, key []byte, iv []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	result := make([]byte, len(encrypter))
	blockMode.CryptBlocks(result, encrypter)
	// 去除填充
	print("result：")
	println(string(result))
	result = pkcs7UnPadding(result)
	return result
}

// ExternalIP 获取公网IP地址
func ExternalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

/*
	PKCS7Padding 填充模式
	text：明文内容
	blockSize：分组块大小
*/
func pkcs7Padding(text []byte, blockSize int) []byte {
	// 计算待填充的长度
	padding := blockSize - len(text)%blockSize
	var paddingText []byte
	if padding == 0 {
		// 已对齐，填充一整块数据，每个数据为 blockSize
		paddingText = bytes.Repeat([]byte{byte(blockSize)}, blockSize)
	} else {
		// 未对齐 填充 padding 个数据，每个数据为 padding
		paddingText = bytes.Repeat([]byte{byte(padding)}, padding)
	}
	return append(text, paddingText...)
}

/*
	去除 PKCS7Padding 填充的数据
	text 待去除填充数据的原文
*/
func pkcs7UnPadding(text []byte) []byte {
	// 取出填充的数据 以此来获得填充数据长度
	unPadding := int(text[len(text)-1])
	return text[:(len(text) - unPadding)]
}
