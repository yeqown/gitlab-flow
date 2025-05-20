package pkg

import (
	"bytes"
	"crypto/des"
	"encoding/base64"
	"fmt"
)

func MustDesDecrypt(ciphertext string, key []byte) string {
	plaintext, err := DesDecrypt(ciphertext, key)
	if err != nil {
		panic(fmt.Sprintf("des decrypt failed: %v", err))
	}

	return string(plaintext)
}

func DesEncrypt(plaintext []byte, key []byte) (string, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS5 填充
	plaintext = PKCS5Padding(plaintext, block.BlockSize())

	// ECB 模式加密
	ciphertext := make([]byte, len(plaintext))
	for bs, be := 0, block.BlockSize(); bs < len(plaintext); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Encrypt(ciphertext[bs:be], plaintext[bs:be])
	}

	// 返回 Base64 编码的字符串
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DesDecrypt(ciphertext string, key []byte) ([]byte, error) {
	// 解码 Base64 字符串
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// ECB 模式解密
	plaintext := make([]byte, len(ciphertextBytes))
	for bs, be := 0, block.BlockSize(); bs < len(ciphertextBytes); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Decrypt(plaintext[bs:be], ciphertextBytes[bs:be])
	}

	// PKCS5 去填充
	plaintext = PKCS5UnPadding(plaintext)
	return plaintext, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
