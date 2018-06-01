package git

import (
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"os"
	"os/exec"
	"strings"
	"time"
)

func LFSFetchAndCheckout(repo string, opts *options.BuildOptions) error {

	if opts.Timings {
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to lfs fetch and dance %s %v\n", opts.Repo, t2)
		}()
	}

	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	err = os.Chdir(repo)

	if err != nil {
		return err
	}

	defer os.Chdir(cwd)

	git_args := make([]string, 0)
	var cmd *exec.Cmd

	git_args = []string{"lfs", "fetch"}
	cmd = exec.Command("git", git_args...)

	opts.Logger.Info("git %s", strings.Join(git_args, " "))

	_, err = cmd.Output()

	if err != nil {
		return err
	}

	git_args = []string{"lfs", "checkout"}
	cmd = exec.Command("git", git_args...)

	opts.Logger.Info("git %s", strings.Join(git_args, " "))

	_, err = cmd.Output()
	return err
}
