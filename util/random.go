package util

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// RandomInt Generate a random integer between min and max
func RandomInt(min, max int64) int64 {
	diff := max - min
	nBig, err := rand.Int(rand.Reader, big.NewInt(diff+1))
	if err != nil {
		panic(err)
	}
	n := nBig.Int64()
	return min + n
}

// RandomString Generate a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := big.NewInt(int64(len(alphabet)))

	for i := 0; i < n; i++ {
		indexBig, err := rand.Int(rand.Reader, k)
		if err != nil {
			panic(err)
		}
		index := indexBig.Int64()
		c := alphabet[index]
		sb.WriteByte(c)
	}
	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

func RandomCurrency() string {
	currencies := []string{USD, EUR, BRL}
	n := big.NewInt(int64(len(currencies)))
	indexBig, err := rand.Int(rand.Reader, n)
	if err != nil {
		panic(err)
	}
	index := indexBig.Int64()
	return currencies[index]
}

func RandomEmail() string {
	return RandomString(6) + "@email.com"
}
