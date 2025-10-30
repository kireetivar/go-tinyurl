package utils

import (
	"crypto/rand"
	"log"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortKey() string {
	s := make([]byte, 6)

	_, err := rand.Read(s)
	if err != nil {
		log.Fatalln("err ", err)
	}

	for i := 0; i < 6; i++ {
		s[i] = alphabet[int(s[i])%len(alphabet)]
	}

	return string(s)
}