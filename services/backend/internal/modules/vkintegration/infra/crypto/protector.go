package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
)

type AESGCMProtector struct {
	aead cipher.AEAD
}

type DisabledProtector struct{}

func NewAESGCMProtector(rawBase64Key string) (*AESGCMProtector, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(rawBase64Key))
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(decoded)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &AESGCMProtector{aead: aead}, nil
}

func NewDisabledProtector() DisabledProtector {
	return DisabledProtector{}
}

func (p *AESGCMProtector) Configured() bool {
	return p != nil && p.aead != nil
}

func (p *AESGCMProtector) Seal(plain string) (string, error) {
	if !p.Configured() {
		return "", vkintegrationusecase.ErrTokenProtectionUnavailable
	}

	nonce := make([]byte, p.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := p.aead.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (p *AESGCMProtector) Open(ciphertext string) (string, error) {
	if !p.Configured() {
		return "", vkintegrationusecase.ErrTokenProtectionUnavailable
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(ciphertext))
	if err != nil {
		return "", vkintegrationusecase.ErrTokenCiphertextCorrupted
	}
	if len(decoded) < p.aead.NonceSize() {
		return "", vkintegrationusecase.ErrTokenCiphertextCorrupted
	}

	nonce := decoded[:p.aead.NonceSize()]
	payload := decoded[p.aead.NonceSize():]

	plain, err := p.aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", vkintegrationusecase.ErrTokenCiphertextCorrupted
	}

	return string(plain), nil
}

func (DisabledProtector) Configured() bool {
	return false
}

func (DisabledProtector) Seal(string) (string, error) {
	return "", vkintegrationusecase.ErrTokenProtectionUnavailable
}

func (DisabledProtector) Open(string) (string, error) {
	return "", vkintegrationusecase.ErrTokenProtectionUnavailable
}
