package utils

import (
	"crypto/rand"
	"log"
	"math/big"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortKey() string {
	const keyLength = 6
	alphabetLen := big.NewInt(int64(len(alphabet)))
	result := make([]byte, keyLength)

	for i := 0; i < keyLength; i++ {
		num, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			log.Fatalln("err ", err)
		}
		result[i] = alphabet[num.Int64()]
	}

	return string(result)
}
