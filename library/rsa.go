package library

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

func RsaEncrypt(origData []byte, publicKey string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)

	returnBytes := []byte{}
	currentIdx := 0
	for ; currentIdx < len(origData); currentIdx += 116 {
		lastIdx := currentIdx + 116
		if lastIdx > len(origData) {
			lastIdx = len(origData)
		}
		currentResult, errEncrypt := rsa.EncryptPKCS1v15(rand.Reader, pub, origData[currentIdx:lastIdx])
		if errEncrypt != nil {
			return nil, err
		}
		returnBytes = append(returnBytes, currentResult...)
	}

	return returnBytes, nil
}

func RsaDecrypt(ciphertext []byte, privateKey string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	returnBytes := []byte{}
	currentIdx := 0
	for ; currentIdx < len(ciphertext); currentIdx += 128 {
		lastIdx := currentIdx + 128
		if lastIdx > len(ciphertext) {
			lastIdx = len(ciphertext)
		}
		currentResult, errDecrypt := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext[currentIdx:lastIdx])
		if errDecrypt != nil {
			return nil, err
		}
		returnBytes = append(returnBytes, currentResult...)
	}

	return returnBytes, nil
}

func RsaEncryptString(origData string, publicKey string) (string, error) {
	outBtye, err := RsaEncrypt([]byte(origData), publicKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(outBtye), nil
}

func RsaDecryptString(ciphertext string, privateKey string) (string, error) {
	rawByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	outByte, err := RsaDecrypt(rawByte, privateKey)
	if err != nil {
		return "", err
	}
	return string(outByte), nil
}
