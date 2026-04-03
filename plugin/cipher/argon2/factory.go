// Package argon2 provides argon2id password hashing functionality.
package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/aawadallak/go-core-kit/core/cipher"

	"golang.org/x/crypto/argon2"
)

const (
	defaultTime    = 1
	defaultMemory  = 64 * 1024
	defaultThreads = 4
	defaultKeyLen  = 32
	defaultSaltLen = 16
)

type Option func(*Adapter)

func WithTime(t uint32) Option {
	return func(a *Adapter) { a.time = t }
}

func WithMemory(m uint32) Option {
	return func(a *Adapter) { a.memory = m }
}

func WithThreads(t uint8) Option {
	return func(a *Adapter) { a.threads = t }
}

func WithKeyLen(k uint32) Option {
	return func(a *Adapter) { a.keyLen = k }
}

type Adapter struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen int
}

func NewAdapter(opts ...Option) cipher.Cipher {
	a := &Adapter{
		time:    defaultTime,
		memory:  defaultMemory,
		threads: defaultThreads,
		keyLen:  defaultKeyLen,
		saltLen: defaultSaltLen,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *Adapter) Encrypt(value string) ([]byte, error) {
	salt := make([]byte, a.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("argon2: generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(value), salt, a.time, a.memory, a.threads, a.keyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		a.memory, a.time, a.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return []byte(encoded), nil
}

func (a *Adapter) Verify(hashedValue []byte, value string) error {
	var version int
	var memory, time uint32
	var threads uint8
	var saltB64, hashB64 string

	n, err := fmt.Sscanf(
		string(hashedValue),
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s",
		&version, &memory, &time, &threads, &saltB64,
	)
	if err != nil || n != 5 {
		return errors.New("argon2: invalid hash format")
	}

	// saltB64 contains "salt$hash", split on $
	parts := splitLast(saltB64, '$')
	if parts == nil {
		return errors.New("argon2: invalid hash format")
	}
	saltB64 = parts[0]
	hashB64 = parts[1]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return fmt.Errorf("argon2: decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return fmt.Errorf("argon2: decode hash: %w", err)
	}

	keyLen := uint32(len(expectedHash)) //nolint:gosec // hash length is bounded by argon2 output, no overflow risk
	computed := argon2.IDKey([]byte(value), salt, time, memory, threads, keyLen)

	if subtle.ConstantTimeCompare(computed, expectedHash) != 1 {
		return cipher.NewErrInvalidHash()
	}

	return nil
}

func splitLast(s string, sep byte) []string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}
