package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
)

func HashFile(path string) (string, error) {

	body, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}

	return HashBytes(body)
}

func HashBytes(body []byte) (string, error) {

	hash := sha256.Sum256(body)
	str := hex.EncodeToString(hash[:])

	return str, nil
}
