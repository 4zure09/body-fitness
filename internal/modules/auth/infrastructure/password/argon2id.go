package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	defaultMemory      uint32 = 64 * 1024
	defaultIterations  uint32 = 3
	defaultParallelism uint8  = 2
	defaultSaltLength         = 16
	defaultKeyLength   uint32 = 32
)

var ErrInvalidHash = errors.New("invalid Argon2id hash")

type Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func New() Hasher {
	return Hasher{
		memory:      defaultMemory,
		iterations:  defaultIterations,
		parallelism: defaultParallelism,
		saltLength:  defaultSaltLength,
		keyLength:   defaultKeyLength,
	}
}

func (h Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	return encodeHash(h, salt, hash), nil
}

func (h Hasher) Verify(encodedHash, password string) (bool, error) {
	parameters, salt, expectedHash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}
	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.iterations,
		parameters.memory,
		parameters.parallelism,
		uint32(len(expectedHash)),
	)
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}

func encodeHash(parameters Hasher, salt, hash []byte) string {
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		parameters.memory,
		parameters.iterations,
		parameters.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func decodeHash(encodedHash string) (Hasher, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Hasher{}, nil, nil, ErrInvalidHash
	}

	version, err := parsePrefixedUint(parts[2], "v=", 8)
	if err != nil || version != argon2.Version {
		return Hasher{}, nil, nil, ErrInvalidHash
	}
	parameters, err := parseParameters(parts[3])
	if err != nil {
		return Hasher{}, nil, nil, err
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return Hasher{}, nil, nil, ErrInvalidHash
	}
	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(hash) < 16 {
		return Hasher{}, nil, nil, ErrInvalidHash
	}
	parameters.saltLength = uint32(len(salt))
	parameters.keyLength = uint32(len(hash))
	return parameters, salt, hash, nil
}

func parseParameters(value string) (Hasher, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 3 {
		return Hasher{}, ErrInvalidHash
	}
	memory, err := parsePrefixedUint(parts[0], "m=", 32)
	if err != nil || memory < 8*1024 || memory > 256*1024 {
		return Hasher{}, ErrInvalidHash
	}
	iterations, err := parsePrefixedUint(parts[1], "t=", 32)
	if err != nil || iterations == 0 || iterations > 10 {
		return Hasher{}, ErrInvalidHash
	}
	parallelism, err := parsePrefixedUint(parts[2], "p=", 8)
	if err != nil || parallelism == 0 || parallelism > 16 {
		return Hasher{}, ErrInvalidHash
	}
	return Hasher{
		memory:      uint32(memory),
		iterations:  uint32(iterations),
		parallelism: uint8(parallelism),
	}, nil
}

func parsePrefixedUint(value, prefix string, bitSize int) (uint64, error) {
	if !strings.HasPrefix(value, prefix) {
		return 0, ErrInvalidHash
	}
	parsed, err := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, bitSize)
	if err != nil {
		return 0, ErrInvalidHash
	}
	return parsed, nil
}
