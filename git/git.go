package git

import (
	"context"
	"errors"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	_ "log"
	"path/filepath"
	"strings"
	"time"
)

type GitTool interface {
	Clone(context.Context, string, string) error
	CommitHashes(...string) (map[string]string, error)
}

func NewGitToolFromOptions(opts *options.BuildOptions) (GitTool, error) {

	var gt GitTool
	var err error

	// the thinking here is that one of two things will happen soon:
	// either the go-git package will support LFS directly or WOF
	// will finish the work to standardize on a default file size that
	// doesn't require LFS - if the latter but not the former happens
	// first then we'll still need to use go-git and shell out to `git lfs`
	// when we are processing large files/repos (hello, NZ) but at least
	// we will have a pure-Go (no shell/exec nonsense) tool for building
	// distributions - today it still requires that there be a git binary
	// (201080816/thisisaaronland)

	switch strings.ToUpper(opts.Cloner) { // it's called Cloner because I haven't updated all that stuff yet (20180823/thisisaaronland)

	case "NATIVE":
		gt, err = NewNativeGitTool()
	default:
		err = errors.New("Invalid Git tool")
	}

	return gt, err
}

func CloneRepo(ctx context.Context, gt GitTool, opts *options.BuildOptions) ([]string, error) {

	if opts.Timings {
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			name, _ := options.DistributionNameFromOptions(opts)
			opts.Logger.Status("time to clone %s %v\n", name, t2)
		}()
	}

	uri := "{protocol}://{source}/{organization}/{repo}.git"

	template, err := uritemplates.Parse(uri)

	if err != nil {
		return nil, err
	}

	local_paths := make([]string, 0)

	// PLEASE DO THIS IN PARALLEL... (20190322/thisisaaronland)

	for _, r := range opts.Repos {

		repo_name := r.Name()

		values := make(map[string]interface{})

		values["protocol"] = opts.Protocol
		values["source"] = opts.Source
		values["organization"] = opts.Organization
		values["repo"] = repo_name

		remote, err := template.Expand(values)

		if err != nil {
			return nil, err
		}

		local := filepath.Join(opts.Workdir, repo_name)

		err = gt.Clone(ctx, remote, local)

		if err != nil {
			return nil, err
		}

		local_paths = append(local_paths, local)
	}

	return local_paths, nil
}
