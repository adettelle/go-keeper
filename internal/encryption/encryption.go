package encryption

/*
func Encrypt(strToEncrypt string, secret string) (string, error) {
	return strToEncrypt, nil
}

func Decrypt(strToDecrypt string, secret string) (string, error) {
	return strToDecrypt, nil
}

*/

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"log"
)

var (
	initialVector = "1234567890123456"
	// passphrase    = "Impassphrasegood"
)

func prepareKey(key []byte) ([]byte, error) {
	if len(key) < 16 {
		return nil, fmt.Errorf("Key is too short")
	}
	lenToGet := len(key) - len(key)%8
	return key[:lenToGet], nil
}

func AESEncrypt(src string, key []byte) (string, error) {
	preparedKey, err := prepareKey(key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(preparedKey)
	if err != nil {
		log.Println("key error:", err)
		return "", err
	}
	if src == "" {
		log.Println("plain content empty:", err)
		return "", err
	}
	ecb := cipher.NewCBCEncrypter(block, []byte(initialVector))
	content := []byte(src)
	content = PKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)

	encryptedString := base64.StdEncoding.EncodeToString(crypted)

	return encryptedString, nil

}

func AESDecrypt(crypt string, key []byte) (string, error) {
	preparedKey, err := prepareKey(key)
	if err != nil {
		return "", err
	}

	encryptedData, err := base64.StdEncoding.DecodeString(crypt)
	if err != nil {
		log.Println("cannot decode base64:", err)
		return "", err
	}

	block, err := aes.NewCipher(preparedKey)
	if err != nil {
		log.Println("key error1:", err)
		return "", err
	}
	if len(encryptedData) == 0 {
		log.Println("plain content empty:", err)
		return "", err
	}
	ecb := cipher.NewCBCDecrypter(block, []byte(initialVector))
	decrypted := make([]byte, len(encryptedData))
	ecb.CryptBlocks(decrypted, encryptedData)

	return string(PKCS5Trimming(decrypted)), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
