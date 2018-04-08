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

func DecPermission(permission int8) (ret ProxyPermission){
	ret.Get = (permission & 0x01) == 0x01
	ret.Get = (permission & 0x02) == 0x02
	ret.Get = (permission & 0x04) == 0x04
	ret.Get = (permission & 0x08) == 0x08
	return ret
}
func EncPermission(input ProxyPermission) (ret int8){
	ret = 0
	if input.Get {
		ret |= 0x01
	}
	if input.Post {
		ret |= 0x02
	}
	if input.Put {
		ret |= 0x04
	}
	if input.Delete {
		ret |= 0x08
	}
	return ret
}

type ProxyPermission struct{
	Get bool
	Post bool
	Put bool
	Delete bool
}