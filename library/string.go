package library

import (
	"math/rand"
)

func GetRandStr(n int) string {
	seedsLetters := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	e := make([]byte, n)
	for i := 0; i < n; i++ {
		e[i] = seedsLetters[rand.Intn(len(seedsLetters))]
	}
	return string(e)
}
