package utils

import (
	"io"
	"os"
)

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
