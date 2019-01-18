package utils

import (
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-repo"
)

func NewRepo(repo_name string, opts *options.BuildOptions) (repo.Repo, error) {

	var r repo.Repo
	var err error

	repo_opts := repo.DefaultFilenameOptions()

	if opts.CustomRepo && opts.LocalCheckout {
		r, err = repo.NewCustomRepoFromPath(repo_name, repo_opts)
	} else if opts.CustomRepo {
		r, err = repo.NewCustomRepoFromString(repo_name)
	} else if opts.LocalCheckout {
		r, err = repo.NewDataRepoFromPath(repo_name, repo_opts)
	} else {
		r, err = repo.NewDataRepoFromString(repo_name)
	}

	return r, err
}
