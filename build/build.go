package build

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-dist/sqlite"
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

	var local_repo string

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

	opts.Logger.Status("LOCAL %s", local_repo)

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return
		default:

			dsn, err := sqlite.BuildSQLite(ctx, local_repo, opts)

			if err != nil {
				err_ch <- err
				return
			}

			opts.Logger.Status("CREATED %s", dsn)
		}
	}

}
