package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"demo520/internal/pkg/log"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

const (
	keyLen  = 32
	saltLen = 32
	time    = 3
	memory  = 64 * 1024
	threads = 4
)

func generateSalt() []byte {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		log.Errorw("Error generating salt", "err", err)
		return nil
	}
	return salt
}

func HashPassword(password string) (string, error) {
	salt := generateSalt()
	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	b64Salt := base64.StdEncoding.EncodeToString(salt)
	b64Hash := base64.StdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s$%s", b64Salt, b64Hash), nil
}

func VerifyPassword(password string, hash string) bool {
	b64Salt, b64Hash := strings.Split(hash, "$")[0], strings.Split(hash, "$")[1]
	curHash, err := base64.StdEncoding.DecodeString(b64Hash)
	if err != nil {
		log.Errorw("Error decoding hash", "err", err)
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(b64Salt)
	if err != nil {
		log.Errorw("Error decoding salt", "err", err)
		return false
	}
	newHash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	if subtle.ConstantTimeCompare(curHash, newHash) == 1 {
		return true
	}
	return false
}
