package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func HashFile(path string) (string, error) {

	fh, err := os.Open(path)

	if err != nil {
		return "", err
	}

	h := sha256.New()

	_, err = io.Copy(h, fh)

	if err != nil {
		return "", err
	}

	hash := h.Sum(nil)
	str := hex.EncodeToString(hash[:])

	return str, nil
}

func HashBytes(body []byte) (string, error) {

	hash := sha256.Sum256(body)
	str := hex.EncodeToString(hash[:])

	return str, nil
}
