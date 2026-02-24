// Package checksum implements the Paytm signature algorithm (SHA256 + AES-CBC).
package checksum

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const iv = "@@@@&&&&####$$$$"

// Generate creates a Paytm checksum signature for the given body using the merchant key.
//
// Algorithm:
//  1. Generate 4 random bytes, hex-encode to get an 8-char salt
//  2. Compute SHA256 of (body + "|" + salt) as hex string
//  3. Concatenate SHA256 hex string + salt
//  4. PKCS7-pad to AES block size
//  5. AES-CBC encrypt with merchant key and fixed IV
//  6. Base64 encode the ciphertext
func Generate(body, merchantKey string) (string, error) {
	salt, err := generateRandomSalt(4)
	if err != nil {
		return "", fmt.Errorf("checksum: failed to generate salt: %w", err)
	}
	return generateWithSalt(body, merchantKey, salt)
}

// Verify verifies a Paytm checksum signature against the given body and merchant key.
func Verify(body, signature, merchantKey string) (bool, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("checksum: invalid base64 signature")
	}

	decrypted, err := aesDecrypt(ciphertext, []byte(merchantKey))
	if err != nil {
		return false, err
	}

	decryptedStr := string(decrypted)

	// decrypted = SHA256_hex(body + "|" + salt) + salt
	// SHA256 hex is always 64 chars
	if len(decryptedStr) <= 64 {
		return false, nil
	}

	storedHash := decryptedStr[:64]
	salt := decryptedStr[64:]

	hash := sha256.Sum256([]byte(body + "|" + salt))
	computedHash := hex.EncodeToString(hash[:])

	return storedHash == computedHash, nil
}

func generateWithSalt(body, merchantKey, salt string) (string, error) {
	hash := sha256.Sum256([]byte(body + "|" + salt))
	hashHex := hex.EncodeToString(hash[:])

	plaintext := []byte(hashHex + salt)

	encrypted, err := aesEncrypt(plaintext, []byte(merchantKey))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func generateRandomSalt(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func aesEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("checksum: failed to create cipher: %w", err)
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, []byte(iv)).CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

func aesDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("checksum: failed to create cipher: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("checksum: ciphertext is not a multiple of the block size")
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, []byte(iv)).CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("checksum: invalid padding")
	}

	return unpadded, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padBytes := make([]byte, padding)
	for i := range padBytes {
		padBytes[i] = byte(padding)
	}
	return append(data, padBytes...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, fmt.Errorf("invalid padding value: %d", padding)
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding byte")
		}
	}
	return data[:len(data)-padding], nil
}
