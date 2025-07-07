package utils

import (
	"crypto/rand"
)

func GenerateOTP() string {
    const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    bytes := make([]byte, 6)
    rand.Read(bytes)
    result := make([]byte, 6)
    for i := range result {
        result[i] = charset[bytes[i]%byte(len(charset))]
    }
    return string(result)
}
