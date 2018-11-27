package compress

import (
	"bytes"
	"fmt"
	"github.com/mholt/archiver"
	_ "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func CompressDirectoryOld(source string, chroot string, opts *CompressOptions) (string, error) {

	abs_source, err := filepath.Abs(source)

	if err != nil {
		return "", err
	}

	dest := fmt.Sprintf("%s.tar.bz2", abs_source)

	abs_chroot, err := filepath.Abs(chroot)

	if err != nil {
		return "", err
	}

	// this bit is important - we are going to tar -C chroot
	// so we don't want to pass an absolute path (to tar)

	rel_source := strings.Replace(source, abs_chroot, "", 1)

	if strings.HasPrefix(rel_source, "/") {
		rel_source = strings.Replace(rel_source, "/", "", 1)
	}

	tar := "tar"

	args := []string{
		"-C", abs_chroot, // -C is for chroot
		"-cjf", // -c is for create; -j is for bzip; -f if for file
		dest,
		rel_source,
	}

	cmd := exec.Command(tar, args...)

	// to do : wire the Logger stuff in to this...

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()

	if err != nil {
		return "", err
	}

	if opts.RemoveSource {

		err = os.RemoveAll(source)

		if err != nil {
			return "", err
		}
	}

	return dest, nil
}
