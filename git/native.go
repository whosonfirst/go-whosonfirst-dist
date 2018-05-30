package git

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"
)

type NativeCloner struct {
	Cloner
	git string
}

func NewNativeCloner() (Cloner, error) {

	// check that git binary is present here...

	cl := NativeCloner{
		git: "git",
	}

	return &cl, nil
}

func (cl *NativeCloner) Clone(ctx context.Context, opts *CloneOptions) (string, error) {

	select {

	case <-ctx.Done():
		return "", nil
	default:

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to clone %s %v\n", opts.Repo, t2)
		}()

		// MAKE THIS CONFIGURABLE

		dir, err := ioutil.TempDir("", opts.Repo)

		if err != nil {
			return "", err
		}

		// DO NOT HOG-TIE THIS TO GITHUB...

		url := fmt.Sprintf("https://github.com/%s/%s.git", opts.Organization, opts.Repo)

		// SOMETHING SOMETHING SOMETHING LFS...

		git_args := []string{
			"clone",
			"--depth",
			"1",
			url,
			dir,
		}

		cmd := exec.Command(cl.git, git_args...)

		_, err = cmd.Output()

		return dir, err
	}
}
