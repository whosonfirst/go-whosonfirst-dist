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
	git string
}

func NewNativeGitTool() (GitTool, error) {

	// check that git binary is present here...

	gt := NativeGitTool{
		git: "git",
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
		// log.Println(gt.git, strings.Join(git_args, " "))

		_, err := cmd.Output()

		return err
	}
}

func (gt *NativeGitTool) CommitHash(path string) (string, error) {

     	cwd, err := os.Getwd()

	if err != nil {
		return "",err
	}

	err = os.Chdir(path)

	if err != nil {
		return "",err
	}

	defer func() {
	      os.Chdir(cwd)
	}()
	
	git_args := []string{
		"log",
		"--pretty=format:'%H'",
		"-n",
		"1",
	}

	cmd := exec.Command(gt.git, git_args...)
	// log.Println(gt.git, strings.Join(git_args, " "))
	
	hash, err := cmd.Output()
	return string(hash), err
}
