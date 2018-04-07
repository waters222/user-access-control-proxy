package util

import (
	"math/rand"
	"crypto/sha1"
	"encoding/hex"
	"crypto/sha256"
)

const randomNumberRune = "0123456789"
const randomCharRune = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(length int, charset string) string{
	b := make([]byte, length)
	chasetLen := len(charset)
	for i:= range b{
		b[i] = charset[rand.Intn(chasetLen)]
	}
	return string(b)
}

func RandomNumberString(length int) string{
	return randomString(length, randomNumberRune)
}
func RandomCharString(length int) string{
	return randomString(length, randomCharRune)
}

// encoding
func EncodeToSha1String(src string) string{
	temp := sha1.Sum([]byte(src))
	return hex.EncodeToString(temp[:])
}
func EncodeToSha256String(src string) string{
	temp := sha256.Sum256([]byte(src))
	return hex.EncodeToString(temp[:])
}