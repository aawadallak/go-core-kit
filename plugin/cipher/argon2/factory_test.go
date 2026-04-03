package argon2

import (
	"errors"
	"strings"
	"testing"

	"github.com/aawadallak/go-core-kit/core/cipher"
)

func TestEncrypt(t *testing.T) {
	a := NewAdapter()

	hash, err := a.Encrypt("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(string(hash), "$argon2id$") {
		t.Errorf("hash should start with $argon2id$, got %q", hash)
	}
}

func TestEncrypt_DifferentSalts(t *testing.T) {
	a := NewAdapter()

	hash1, _ := a.Encrypt("password123")
	hash2, _ := a.Encrypt("password123")

	if string(hash1) == string(hash2) {
		t.Error("two hashes of the same password should differ due to random salt")
	}
}

func TestVerify_CorrectPassword(t *testing.T) {
	a := NewAdapter()

	hash, err := a.Encrypt("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := a.Verify(hash, "password123"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	a := NewAdapter()

	hash, err := a.Encrypt("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = a.Verify(hash, "wrongpassword")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, cipher.ErrInvalidHash) {
		t.Errorf("expected ErrInvalidHash, got %v", err)
	}
}

func TestVerify_InvalidHashFormat(t *testing.T) {
	a := NewAdapter()

	err := a.Verify([]byte("not-a-valid-hash"), "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewAdapter_WithOptions(t *testing.T) {
	a := NewAdapter(
		WithTime(2),
		WithMemory(32*1024),
		WithThreads(2),
		WithKeyLen(64),
	)

	hash, err := a.Encrypt("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := a.Verify(hash, "password123"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAdapter_ImplementsCipherInterface(t *testing.T) {
	var _ cipher.Cipher = NewAdapter()
}
