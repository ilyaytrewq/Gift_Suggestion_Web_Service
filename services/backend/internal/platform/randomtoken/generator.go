package randomtoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

type Generator struct {
	size int
}

func NewGenerator(size int) *Generator {
	return &Generator{size: size}
}

func (g *Generator) NewToken() (string, string, error) {
	buffer := make([]byte, g.size)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", err
	}

	rawToken := base64.RawURLEncoding.EncodeToString(buffer)

	return rawToken, Hash(rawToken), nil
}

func (g *Generator) Hash(rawToken string) string {
	return Hash(rawToken)
}

func Hash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
