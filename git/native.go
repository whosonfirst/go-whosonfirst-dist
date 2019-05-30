package git

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
)

type NativeGitTool struct {
	GitTool
	git   string
	Debug bool
}

func NewNativeGitTool() (GitTool, error) {

	// check that git binary is present here...

	gt := NativeGitTool{
		git:   "git",
		Debug: false,
	}

	return &gt, nil
}

func (gt *NativeGitTool) Clone(ctx context.Context, remote string, local string) error {

	select {

	case <-ctx.Done():
		return nil
	default:

		git_args := []string{
			"lfs",
			"clone",
			"--depth",
			"1",
			remote,
			local,
		}

		cmd := exec.Command(gt.git, git_args...)

		if gt.Debug {
			log.Println(gt.git, strings.Join(git_args, " "))
		}

		_, err := cmd.Output()

		return err
	}
}

func (gt *NativeGitTool) CommitHash(paths ...string) (string, error) {

	type HashResponse struct {
		Index int
		Hash  string
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)
	hash_ch := make(chan HashResponse)

	hash_path := func(ctx context.Context, path string, idx int) {

		defer func() {
			done_ch <- true
		}()

		select {
		case <-ctx.Done():
			return
		default:
			// pass
		}

		cwd, err := os.Getwd()

		if err != nil {
			err_ch <- err
			return
		}

		err = os.Chdir(path)

		if err != nil {
			err_ch <- err
			return
		}

		defer func() {
			os.Chdir(cwd)
		}()

		git_args := []string{
			"log",
			"--pretty=format:%H",
			"-n",
			"1",
		}

		cmd := exec.Command(gt.git, git_args...)

		if gt.Debug {
			log.Println(gt.git, strings.Join(git_args, " "))
		}

		hash, err := cmd.Output()

		if err != nil {
			err_ch <- err
			return
		}

		rsp := HashResponse{
			Index: idx,
			Hash:  string(hash),
		}

		hash_ch <- rsp
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for idx, path := range paths {
		go hash_path(ctx, path, idx)
	}

	remaining := len(paths)
	hashes := make([]string, remaining)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return "", err
		case hash_rsp := <-hash_ch:
			hashes[hash_rsp.Index] = hash_rsp.Hash
		}
	}

	return strings.Join(hashes, ":"), nil
}
