package build

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-dist/bundles"
	"github.com/whosonfirst/go-whosonfirst-dist/csv"
	"github.com/whosonfirst/go-whosonfirst-dist/database"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"os"
	"time"
)

func BuildDistributions(opts *options.BuildOptions, repos []string) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build all %v\n", t2)
		}()
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, repo := range repos {

		local_opts := opts.Clone()
		local_opts.Repo = repo

		go BuildDistribution(ctx, local_opts, done_ch, err_ch)
	}

	count := len(repos)

	for count > 0 {

		select {
		case <-done_ch:
			count--
		case err := <-err_ch:

			if opts.Strict {
				// remember we're defer cancel() -ing above
				return err
			}

			opts.Logger.Error("%v", err)

		default:
			// pass
		}
	}

	return nil
}

func BuildDistribution(ctx context.Context, opts *options.BuildOptions, done_ch chan bool, err_ch chan error) {

	if opts.Timings {
		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build %s %v\n", opts.Repo, t2)
		}()
	}

	defer func() {
		done_ch <- true
	}()

	// eventually these will all be replaced by distibution Item and/or
	// distribution.Inventory thingies... (20180613/thisisaaronland)

	var local_repo string
	var local_sqlite string
	var local_metafiles []string
	var local_bundlefiles []string

	select {

	case <-ctx.Done():
		return
	default:

		if !opts.LocalCheckout {

			repo, err := git.CloneRepo(ctx, opts)

			if err != nil {
				err_ch <- err
				return
			}

			defer func() {

				if opts.PreserveCheckout {
					opts.Logger.Info("local checkout left in place at %s", repo)
				} else {
					os.RemoveAll(repo)
				}
			}()

			local_repo = repo

		} else {
			local_repo = opts.Repo
		}

	}

	opts.Logger.Status("local_repo is %s", local_repo)

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return
		default:

			dsn, err := database.BuildSQLite(ctx, local_repo, opts)

			if err != nil {
				err_ch <- err
				return
			}

			local_sqlite = dsn

			opts.Logger.Status("local sqlite is %s", local_sqlite)
		}
	}

	// eventually we should be able to do these two operations in parallel
	// assuming that a SQLite database has been created - and the bundles
	// code has been updated to read from those databases...
	// (20180602/thisisaaronland)

	if opts.Meta {

		mode := "repo"
		source := local_repo

		if opts.SQLite {
			mode = "sqlite"
			source = local_sqlite
		}

		opts.Logger.Status("build metafile from %s (%s)", mode, source)

		select {

		case <-ctx.Done():
			return
		default:

			metafiles, err := csv.BuildMetaFiles(ctx, opts, mode, source)

			if err != nil {
				err_ch <- err
				return
			}

			local_metafiles = metafiles
			opts.Logger.Status("built metafiles %s", local_metafiles)
		}
	}

	if opts.Bundle {

		select {

		case <-ctx.Done():
			return
		default:

			source := local_repo // PLEASE UPDATE ME READ FROM sqlite ALSO

			bundlefiles, err := bundles.BuildBundle(ctx, local_metafiles, source)

			if err != nil {
				err_ch <- err
				return
			}

			local_bundlefiles = bundlefiles
			opts.Logger.Debug("%v", local_bundlefiles) // temporary - just to make the compiler shut up...
		}

	}

}
