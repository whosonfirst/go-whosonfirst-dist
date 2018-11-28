package compress

import (
	"fmt"
	"github.com/mholt/archiver"
	_ "log"
	"path/filepath"
)

// https://godoc.org/github.com/mholt/archiver#TarBz2
// https://godoc.org/github.com/mholt/archiver#Tar

func CompressDirectory(source string, chroot string, opts *CompressOptions) (string, error) {

	path_source, err := filepath.Abs(source)

	if err != nil {
		return "", err
	}

	path_dest := fmt.Sprintf("%s.tar.bz2", path_source)

	if err != nil {
		return "", err
	}

	// it is unclear to me whether we need to do a chroot dance
	// (and back) here... tbd (20181127/thisisaaronland)

	arch := archiver.NewTarBz2()

	err = arch.Archive([]string{path_source}, path_dest)

	if err != nil {
		return "", err
	}

	return path_dest, nil
}
