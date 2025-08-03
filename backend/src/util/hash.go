package util

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// CalculateSHA256 calculates the SHA256 hash of an io.Reader.
func CalculateSHA256(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to calculate SHA256 hash: %w", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
