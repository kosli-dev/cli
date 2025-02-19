package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	keyring "github.com/zalando/go-keyring"
)

const CredentialsStoreService = "KosliCLIService"

func GenerateRandomAESKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256 requires a 32-byte key
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func AESEncrypt(input string, key []byte) ([]byte, error) {
	inputBytes := []byte(input)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(inputBytes))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], inputBytes)

	return ciphertext, nil
}

func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

func GetSecretFromCredentialsStore(secretName string) (string, error) {
	secret, err := keyring.Get(CredentialsStoreService, secretName)
	if err != nil {
		return "", err
	}
	return secret, nil
}

func SetSecretInCredentialsStore(secretName, secretValue string) error {
	return keyring.Set(CredentialsStoreService, secretName, secretValue)
}
