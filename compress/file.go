package compress

import (
	"fmt"
	"github.com/mholt/archiver/v3"	
	"github.com/whosonfirst/atomicfile"
	"os"
	"path/filepath"
)

// https://godoc.org/github.com/mholt/archiver#Bz2

func CompressFile(source string, chroot string, opts *CompressOptions) (string, error) {

	path_source, err := filepath.Abs(source)

	if err != nil {
		return "", err
	}

	path_dest := fmt.Sprintf("%s.bz2", path_source)

	if err != nil {
		return "", err
	}

	in, err := os.Open(path_source)

	if err != nil {
		return "", err
	}

	defer in.Close()

	out, err := atomicfile.New(path_dest, 0644)

	if err != nil {
		return "", err
	}

	arch := archiver.NewBz2()

	err = arch.Compress(in, out)

	if err != nil {

		err = out.Abort()

		if err != nil {
			return "", err
		}
	}

	err = out.Close()

	if err != nil {
		return "", err
	}

	return path_dest, nil
}
