package crypto

import (
	"errors"
	"testing"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
)

const testVKEncryptionKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func TestAESGCMProtectorRoundTrip(t *testing.T) {
	t.Parallel()

	protector, err := NewAESGCMProtector(testVKEncryptionKey)
	if err != nil {
		t.Fatalf("NewAESGCMProtector() error = %v", err)
	}

	sealed, err := protector.Seal("secret-token")
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	if sealed == "secret-token" {
		t.Fatal("Seal() returned plain text")
	}

	plain, err := protector.Open(sealed)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if plain != "secret-token" {
		t.Fatalf("Open() = %q, want %q", plain, "secret-token")
	}
}

func TestAESGCMProtectorRejectsCorruptedCiphertext(t *testing.T) {
	t.Parallel()

	protector, err := NewAESGCMProtector(testVKEncryptionKey)
	if err != nil {
		t.Fatalf("NewAESGCMProtector() error = %v", err)
	}

	_, err = protector.Open("not-base64")
	if !errors.Is(err, vkintegrationusecase.ErrTokenCiphertextCorrupted) {
		t.Fatalf("Open() error = %v, want %v", err, vkintegrationusecase.ErrTokenCiphertextCorrupted)
	}
}

func TestDisabledProtectorRejectsSealAndOpen(t *testing.T) {
	t.Parallel()

	protector := NewDisabledProtector()
	if protector.Configured() {
		t.Fatal("Configured() = true, want false")
	}

	if _, err := protector.Seal("secret"); !errors.Is(err, vkintegrationusecase.ErrTokenProtectionUnavailable) {
		t.Fatalf("Seal() error = %v, want %v", err, vkintegrationusecase.ErrTokenProtectionUnavailable)
	}
	if _, err := protector.Open("secret"); !errors.Is(err, vkintegrationusecase.ErrTokenProtectionUnavailable) {
		t.Fatalf("Open() error = %v, want %v", err, vkintegrationusecase.ErrTokenProtectionUnavailable)
	}
}
