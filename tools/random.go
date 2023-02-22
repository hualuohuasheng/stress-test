package tools

import (
	"math/rand"
	"strconv"
	"time"
)

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStrByTimestamp() string {
	return strconv.Itoa(int(time.Now().UnixNano() / 1000))
}

func RandStrByLetters(n int) string {
	b := make([]rune, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(62)]
	}
	return string(b)
}
