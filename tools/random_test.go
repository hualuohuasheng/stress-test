package tools

import (
	"testing"
)

func TestAESEncryptData(t *testing.T) {
	enData := AESEncryptData("5f20e76fe8ba4b6cbb0d316fb3478f2c", "123456", "abcd1234abcd", "")
	t.Log(enData)
	enData = AESEncryptData("5f20e76fe8ba4b6cbb0d316fb3478f2c", "123456", "abcd1234abcd", "1234")
	//fmt.Println(enData)
	t.Log(enData)
}

func TestAESDecryptData(t *testing.T) {
	deData := AESDecryptData("5f20e76fe8ba4b6cbb0d316fb3478f2c", "abcd1234abcd", "", "T8TrkS7CqQpp0sgcHbqRG/k2IynDEA==")
	t.Log(deData)
	deData = AESDecryptData("5f20e76fe8ba4b6cbb0d316fb3478f2c", "abcd1234abcd", "1234", "T8TrkS7CeyZmvH8t++0yI/jQ1+jIdw==")
	t.Log(deData)
	if deData != "123456" {
		t.Fatal(deData)
	}
}
