package tools

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

func ReadKeyFromPermFile(permFile string) []byte {
	keyFile, err := os.Open(permFile)
	if err != nil {
		fmt.Println("open file Error: ", err)
		panic(err)
	}
	defer func() {
		_ = keyFile.Close()
	}()
	fileState, _ := keyFile.Stat()
	buf := make([]byte, fileState.Size())
	_, err = keyFile.Read(buf)
	if err != nil {
		fmt.Println("read perm file Error: ", err)
		panic(err)
	}
	//fmt.Println("文件大小: ", fileSize)
	return buf
}

func RsaEncryptUsePrivateKey(keyBuf []byte, plaintext string) string {
	block, _ := pem.Decode(keyBuf)
	privateKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	msg := []byte(plaintext)
	msgHash := sha256.New()
	_, err := msgHash.Write(msg)
	if err != nil {
		panic(err)
	}
	msgHashSum := msgHash.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHashSum)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(signature)
}

func RsaDecryptUsePublicKey(keyBuf []byte, plaintext, ciphertext string) bool {
	block, _ := pem.Decode(keyBuf)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	publicKey := pubInterface.(*rsa.PublicKey)

	msgHash := sha256.New()
	_, err := msgHash.Write([]byte(plaintext))
	if err != nil {
		panic(err)
	}
	msgHashSum := msgHash.Sum(nil)

	cipherData, _ := base64.StdEncoding.DecodeString(ciphertext)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, msgHashSum, cipherData)
	if err != nil {
		return false
	} else {
		return true
	}
}

func AESEncryptData(secKey, plaintext, nonce, addData string) string {
	key := []byte(secKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	aesGcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	var ciphertext []byte
	if addData != "" {
		ciphertext = aesGcm.Seal(nil, []byte(nonce), []byte(plaintext), []byte(addData))
	} else {
		ciphertext = aesGcm.Seal(nil, []byte(nonce), []byte(plaintext), nil)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func AESDecryptData(secKey, nonce, addData, cipherText string) string {
	cipherData, _ := base64.StdEncoding.DecodeString(cipherText)

	block, err := aes.NewCipher([]byte(secKey))
	if err != nil {
		panic(err.Error())
	}

	aesGcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		panic(err.Error())
	}
	var plaintext []byte
	if addData != "" {
		plaintext, err = aesGcm.Open(nil, []byte(nonce), cipherData, []byte(addData))
	} else {
		plaintext, err = aesGcm.Open(nil, []byte(nonce), cipherData, nil)
	}
	return string(plaintext)
}
