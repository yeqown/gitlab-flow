package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Des(t *testing.T) {
	// 示例使用
	plaintext := []byte("Hello, DES!")
	key := []byte("12345678") // DES 密钥长度必须为 8 字节

	// 加密
	ciphertext, err := DesEncrypt(plaintext, key)
	assert.NoError(t, err)
	assert.Equal(t, "fy7lEyzVC0iQwyShwfm6Vg==", ciphertext)

	// 解密
	decrypted, err := DesDecrypt(ciphertext, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
