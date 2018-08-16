package git

import (
	"context"
	"errors"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"path/filepath"
	"time"
)

type Cloner interface {
	Clone(context.Context, string, string) error
}

func NewClonerFromOptions(opts *options.BuildOptions) (Cloner, error) {

	var cl Cloner
	var err error

	switch opts.Cloner {

	// the thinking here is that one of two things will happen soon:
	// either the go-git package will support LFS directly or WOF
	// will finish the work to standardize on a default file size that
	// doesn't require LFS - if the latter but not the former happens
	// first then we'll still need to use go-git and shell out to `git lfs`
	// when we are processing large files/repos (hello, NZ) but at least
	// we will have a pure-Go (no shell/exec nonsense) tool for building
	// distributions - today it still requires that there be a git binary
	// (201080816/thisisaaronland)

	case "native":
		cl, err = NewNativeCloner()
	default:
		err = errors.New("Invalid cloner")
	}

	return cl, err
}

func CloneRepo(ctx context.Context, opts *options.BuildOptions) (string, error) {

	if opts.Timings {
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to clone %s %v\n", opts.Repo, t2)
		}()
	}

	uri := "{protocol}://{source}/{organization}/{repo}.git"

	template, err := uritemplates.Parse(uri)

	if err != nil {
		return "", err
	}

	repo_name := opts.Repo.Name()

	values := make(map[string]interface{})

	values["protocol"] = opts.Protocol
	values["source"] = opts.Source
	values["organization"] = opts.Organization
	values["repo"] = repo_name

	remote, err := template.Expand(values)

	if err != nil {
		return "", err
	}

	local := filepath.Join(opts.Workdir, repo_name)

	// SOMETHING SOMETHING SOMETHING check for presence of git-lfs
	// (20180604/thisisaaronland)

	cl, err := NewClonerFromOptions(opts)

	if err != nil {
		return "", err
	}

	err = cl.Clone(ctx, remote, local)

	if err != nil {
		return "", err
	}

	return local, nil
}
