package utils

import (
	"github.com/whosonfirst/go-whosonfirst-dist/compress"
	"github.com/whosonfirst/go-whosonfirst-dist/hash"
	"io"
	"os"
	"path/filepath"
)

func CompressFile(source string) (string, string, error) {

	opts := compress.DefaultCompressOptions()
	root := filepath.Dir(source)

	path_compressed, err := compress.CompressFile(source, root, opts)

	if err != nil {
		return "", "", err
	}

	hashed, err := hash.HashFile(path_compressed)

	if err != nil {
		os.Remove(path_compressed)
		return "", "", err
	}

	return path_compressed, hashed, nil
}

// because os.Rename can't across devices on Linux...
// (20180604/thisisaaronland)

func Rename(src string, dest string) error {

	info, err := os.Stat(src)

	if err != nil {
		return err
	}

	in, err := os.Open(src)

	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, info.Mode())

	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)

	out.Close()
	return err
}
