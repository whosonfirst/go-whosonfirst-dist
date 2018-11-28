package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
)

/*

or this:

shasum -a 256 src-whosonfirst-data-latest.db
d515bee2e63609f9e9bf9c91dd5e7464debadd9fcd46cadf4358c7150ef85a1f  src-whosonfirst-data-latest.db

*/

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

func HashFileOld(path string) (string, error) {

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
