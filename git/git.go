package git

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io/ioutil"
	"time"
)

type Cloner interface {
	Clone(context.Context, string, string) error
}

type CloneOptions struct {
	Cloner       Cloner
	Organization string
	Repo         string
	Logger       *log.WOFLogger
}

func CloneRepo(ctx context.Context, opts *CloneOptions) (string, error) {

	t1 := time.Now()

	defer func() {
		t2 := time.Since(t1)
		opts.Logger.Status("time to clone %s %v\n", opts.Repo, t2)
	}()

	// MAKE THIS CONFIGURABLE

	local, err := ioutil.TempDir("", opts.Repo)

	if err != nil {
		return "", err
	}

	// DO NOT HOG-TIE THIS TO GITHUB...

	remote := fmt.Sprintf("https://github.com/%s/%s.git", opts.Organization, opts.Repo)

	// SOMETHING SOMETHING LFS

	cl := opts.Cloner

	err = cl.Clone(ctx, local, remote)
	return local, err
}
