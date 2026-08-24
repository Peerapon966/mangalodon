package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/Peerapon966/blackbox/scraper/internal/apperr"
)

// Encrypt payload using AES-256-GCM.
// The output must strictly be: [12-byte IV] + [Ciphertext] + [16-byte Auth Tag]
// to match the exact format expected by Web Crypto API in the React frontend.
func EncryptBlob(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.EncryptionFailed,
			Message: "failed to create AES cipher block",
			Err:     err.Error(),
		}
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.EncryptionFailed,
			Message: "failed to create GCM cipher mode",
			Err:     err.Error(),
		}
	}

	// Web Crypto requires a 12-byte IV for AES-GCM
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.EncryptionFailed,
			Message: "failed to generate 12-byte random IV",
			Err:     err.Error(),
		}
	}

	// Seal appends the auth tag automatically to the end of the ciphertext
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	// Combine IV + Ciphertext (which includes auth tag at the end)
	result := append(nonce, ciphertext...)

	return result, nil
}

// DecryptBlob decrypts AES-256-GCM payloads.
// It expects the first 12 bytes to be the IV, followed by the ciphertext and the 16-byte auth tag.
func DecryptBlob(encrypted []byte, key []byte) ([]byte, error) {
	if len(encrypted) < 12 {
		return nil, &apperr.ScraperError{
			Code:    apperr.DecryptionFailed,
			Message: "ciphertext too short to contain IV",
		}
	}

	iv := encrypted[:12]
	ciphertext := encrypted[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.DecryptionFailed,
			Message: "failed to create AES cipher block",
			Err:     err.Error(),
		}
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.DecryptionFailed,
			Message: "failed to create GCM cipher mode",
			Err:     err.Error(),
		}
	}

	// Open automatically verifies the auth tag at the end of the ciphertext
	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, &apperr.ScraperError{
			Code:    apperr.DecryptionFailed,
			Message: "failed to decrypt payload or authenticate tag",
			Err:     err.Error(),
		}
	}

	return plaintext, nil
}
