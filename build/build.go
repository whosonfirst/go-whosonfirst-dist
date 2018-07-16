package build

import (
	_ "compress/bzip2"
	"context"
	"errors"
	"fmt"
	_ "github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/csv"
	"github.com/whosonfirst/go-whosonfirst-dist/database"
	"github.com/whosonfirst/go-whosonfirst-dist/fs"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	_ "io"
	_ "net/http"
	"os"
	"path/filepath"
	"time"
)

func BuildDistributions(opts *options.BuildOptions, repos []string) ([]dist.Distribution, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build all %v\n", t2)
		}()
	}

	dist_ch := make(chan dist.Distribution)
	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, repo := range repos {

		local_opts := opts.Clone()
		local_opts.Repo = repo

		go BuildDistribution(ctx, local_opts, dist_ch, done_ch, err_ch)
	}

	var build_err error

	build_items := make([]dist.Distribution, 0)
	count := len(repos)

	for count > 0 {

		select {
		case <-done_ch:
			count--
		case d := <-dist_ch:
			build_items = append(build_items, d)
		case err := <-err_ch:

			opts.Logger.Error("%v", err)
			build_err = err

			// remember we're defer cancel() -ing above

			if opts.Strict {
				break
			}

		default:
			// pass
		}
	}

	return build_items, build_err
}

func BuildDistribution(ctx context.Context, opts *options.BuildOptions, dist_ch chan dist.Distribution, done_ch chan bool, err_ch chan error) {

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

	// do we need to work with a remote (or local) Git checkout and if so
	// where is it?

	select {

	case <-ctx.Done():
		return
	default:

		if opts.LocalCheckout || opts.LocalSQLite {
			local_repo = opts.Repo
		} else {

			// SOMETHING SOMETHING throw an error if local_repo exists or remove?
			// (20181013/thisisaaronland)

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
		}
	}

	opts.Logger.Status("local_repo is %s", local_repo)

	// if opts.RemoteSQLite then fetch from dist.whosonfirst.org (and uncompressed) and
	// store in opts.Workdir here... (20180704/thisisaaronland)

	/*
		if opts.RemoteSQLite {

			local_fname := fmt.Sprintf("%s-latest.db", opts.Repo)
			local_sqlite = filepath.Join(opts.Workdir, local_fname)

			remote_fname := fmt.Sprintf("%s.bz2", local_fname)
			remote_sqlite := fmt.Sprintf("https://dist.whosonfirst.org/sqlite/%s", remote_fname)

			rsp, err := http.Get(remote_sqlite)

			if err != nil {
				err_ch <- err
				return
			}

			defer rsp.Body.Close()

			br := bzip2.NewReader(rsp.Body)

			wr, err := atomicfile.New(local_sqlite, 0644)

			if err != nil {
				err_ch <- err
				return
			}

			_, err = io.Copy(wr, br)

			if err != nil {
				wr.Abort()

				err_ch <- err
				return
			}

			err = wr.Close()

			if err != nil {
				err_ch <- err
				return
			}

			logger.Info("Retrieved remote SQLite (%s) and stored as %s", remote_sqlite, local_sqlite)
		}
	*/

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return
		default:

			// PLEASE FIX ME
			// 1. use go-whosonfirst-repo
			// 2. reconcile me with the code in database/sqlite.go
			//    which should also do (1)
			// (20180620/thisisaaronland)

			if opts.LocalSQLite {

				fname := fmt.Sprintf("%s-latest.db", opts.Repo)
				local_sqlite = filepath.Join(opts.Workdir, fname)

			} else {

				d, err := database.BuildSQLite(ctx, local_repo, opts)

				if err != nil {
					err_ch <- err
					return
				}

				// should we do this now?
				dist_ch <- d

				local_sqlite = d.Path()
			}

			opts.Logger.Status("local sqlite is %s", local_sqlite)
		}
	}

	_, err := os.Stat(local_sqlite)

	if err != nil {
		err_ch <- err
		return
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

			d_many, err := csv.BuildMetaFiles(ctx, opts, mode, source)

			if err != nil {
				err_ch <- err
				return
			}

			if len(d_many) == 0 {
				err_ch <- errors.New("No metafiles produced")
				return
			}

			for _, d := range d_many {
				dist_ch <- d
				local_metafiles = append(local_metafiles, d.Path())
			}

			opts.Logger.Status("built metafiles %s", local_metafiles)
		}
	}

	if opts.Bundle {

		select {

		case <-ctx.Done():
			return
		default:

			// SOMETHING SOMETHING throw an error if local_bundlefiles exist or remove?
			// That presumes knowing what they are called first and/or moving this check
			// in to fs.BuildBundle...
			// (20181013/thisisaaronland)

			bundlefiles, err := fs.BuildBundle(ctx, opts, local_metafiles, local_sqlite)

			if err != nil {
				err_ch <- err
				return
			}

			local_bundlefiles = bundlefiles

			opts.Logger.Status("made bundle %s", local_bundlefiles)
		}
	}

}
