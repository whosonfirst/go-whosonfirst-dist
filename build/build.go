package build

import (
	"context"
	"errors"
	"github.com/whosonfirst/go-whosonfirst-dist"
	"github.com/whosonfirst/go-whosonfirst-dist/csv"
	"github.com/whosonfirst/go-whosonfirst-dist/database"
	"github.com/whosonfirst/go-whosonfirst-dist/fs"
	"github.com/whosonfirst/go-whosonfirst-dist/git"
	"github.com/whosonfirst/go-whosonfirst-dist/options"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func BuildDistributionsForRepos(ctx context.Context, opts *options.BuildOptions, repos ...repo.Repo) (map[string][]*dist.Item, error) {

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build distributions for %d repos %v", len(repos), t2)
		}()
	}

	type BuildItem struct {
		Repo  repo.Repo
		Items []*dist.Item
	}

	build_ch := make(chan BuildItem)
	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, r := range repos {

		go func(ctx context.Context, r repo.Repo, build_ch chan BuildItem, done_ch chan bool, err_ch chan error) {

			defer func() {
				done_ch <- true
			}()

			local_opts := opts.Clone()
			local_opts.Repo = r

			items, err := BuildDistributions(ctx, local_opts)

			opts.Logger.Status("build for %s : %v", r.String(), err)

			if err != nil {
				err_ch <- err
				return
			}

			b := BuildItem{
				Repo:  r,
				Items: items,
			}

			build_ch <- b

		}(ctx, r, build_ch, done_ch, err_ch)
	}

	items := make(map[string][]*dist.Item)
	var err error

	remaining := len(repos)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining--
		case b := <-build_ch:
			items[b.Repo.Name()] = b.Items
		case e := <-err_ch:
			err = e
			break
		default:
			// pass
		}
	}

	if err != nil {
		return nil, err
	}

	return items, nil
}

func BuildDistributions(ctx context.Context, opts *options.BuildOptions) ([]*dist.Item, error) {

	distributions := make([]dist.Distribution, 0) // uncompressed and private
	items := make([]*dist.Item, 0)                // compressed and public

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build COMPRESSED distributions for %s %v", opts.Repo, t2)
		}()
	}

	defer func() {

		if opts.Timings {

			t1 := time.Now()

			defer func() {
				t2 := time.Since(t1)
				opts.Logger.Status("time to remove uncompressed files for %s %v", opts.Repo, t2)
			}()
		}

		rm := func(path string) {

			opts.Logger.Status("remove uncompressed file %s", path)

			info, err := os.Stat(path)

			if os.IsNotExist(err) {
				return
			}

			if err != nil {
				opts.Logger.Warning("Failed to stat path '%s' : %s", path, err)
				return
			}

			if info.IsDir() {
				err = os.RemoveAll(path)
			} else {
				err = os.Remove(path)
			}

			if err != nil {
				opts.Logger.Warning("Failed to remove '%s' : %s", path, err)
			}
		}

		wg := new(sync.WaitGroup)

		for _, d := range distributions {

			t := d.Type()

			switch t.Class() {

			case "csv":

				if t.Major() == "meta" && opts.PreserveMeta {
					continue
				}

			case "database":

				if t.Major() == "sqlite" && opts.PreserveSQLite {
					continue
				}

			case "fs":

				if t.Major() == "bundle" && opts.PreserveBundle {
					continue
				}

			default:
				// pass
			}

			wg.Add(1)

			go func(path string, wg *sync.WaitGroup) {

				defer wg.Done()
				rm(path)

			}(d.Path(), wg)
		}

		wg.Wait()
	}()

	distributions, meta, err := buildDistributionsForRepo(ctx, opts)

	if err != nil {
		opts.Logger.Warning("build (buildDistributionsForRepo) for repo %s failed: %s", opts.Repo, err)
		return nil, err
	}

	item_ch := make(chan *dist.Item)
	done_ch := make(chan bool)
	err_ch := make(chan error)

	// something something something size of the file/directory?

	count_throttle := opts.CompressMaxCPUs

	throttle_ch := make(chan bool, count_throttle)

	for i := 0; i < count_throttle; i++ {
		throttle_ch <- true
	}

	for _, d := range distributions {

		go func(ctx context.Context, d dist.Distribution, item_ch chan *dist.Item, throttle_ch chan bool, done_ch chan bool, err_ch chan error) {

			defer func() {
				opts.Logger.Status("All done compressing %s", d.Path())
				done_ch <- true
			}()

			opts.Logger.Status("register function to compress %s", d.Path())

			if opts.Timings {

				t1 := time.Now()

				defer func() {
					t2 := time.Since(t1)
					opts.Logger.Status("time to compress %s %v", d.Path(), t2)
				}()
			}

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			ta := time.Now()

			<-throttle_ch

			if opts.Timings {
				tb := time.Since(ta)
				opts.Logger.Status("time to wait to start compressing %s %v", d.Path(), tb)
			}

			defer func() {
				opts.Logger.Status("All done compressing %s (throttle)", d.Path())
				throttle_ch <- true
			}()

			c, err := d.Compress()

			if err != nil {
				err_ch <- err
				return
			}

			i, err := dist.NewItemFromDistribution(d, c, meta)

			if err != nil {
				err_ch <- err
				return
			}

			item_ch <- i

		}(ctx, d, item_ch, throttle_ch, done_ch, err_ch)
	}

	var item_err error
	remaining := len(distributions)

	for remaining > 0 {

		select {

		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			item_err = e
			break
		case i := <-item_ch:
			items = append(items, i)
		default:
			// pass
		}
	}

	if item_err != nil {
		return nil, item_err
	}

	return items, nil
}

func buildDistributionsForRepo(ctx context.Context, opts *options.BuildOptions) ([]dist.Distribution, *dist.MetaData, error) {

	if opts.Timings {

		t1 := time.Now()

		defer func() {
			t2 := time.Since(t1)
			opts.Logger.Status("time to build UNCOMPRESSED distributions for %s %v", opts.Repo, t2)
		}()
	}

	distributions := make([]dist.Distribution, 0)

	select {

	case <-ctx.Done():
		return nil, nil, nil
	default:
		// pass
	}

	var local_checkout string
	var local_sqlite string
	var local_metafiles []string
	var local_bundlefiles []string

	gt, err := git.NewGitToolFromOptions(opts)

	if err != nil {
		return nil, nil, err
	}

	// do we need to work with a remote (or local) Git checkout and if so
	// where is it?

	if opts.LocalCheckout || opts.LocalSQLite {
		local_checkout = filepath.Join(opts.Workdir, opts.Repo.Name())
	} else {

		// SOMETHING SOMETHING throw an error if local_checkout exists or remove?
		// (20181013/thisisaaronland)

		repo_path, err := git.CloneRepo(ctx, gt, opts)

		if err != nil {
			return nil, nil, err
		}

		local_checkout = repo_path
	}

	// I don't love that this is here...

	if !opts.PreserveCheckout {

		defer func() {

			err := os.RemoveAll(local_checkout)

			if err != nil {
				opts.Logger.Status("failed to remove %s, %s", local_checkout, err)
			}
		}()
	}

	opts.Logger.Status("local_checkout is %s", local_checkout)

	commit_hash, err := gt.CommitHash(local_checkout)

	if err != nil {
		opts.Logger.Warning("failed to determine commit hash for %s, %s", local_checkout, err)
		commit_hash = ""
	} else {
		opts.Logger.Status("commit hash is %s (%s)", commit_hash, local_checkout)
	}

	// if opts.RemoteSQLite then fetch from dist.whosonfirst.org (and uncompressed) and
	// store in opts.Workdir here... (20180704/thisisaaronland)

	/*
		if opts.RemoteSQLite {

			local_fname := opts.Repo.SQLiteFilename()	// fmt.Sprintf("%s-latest.db", opts.Repo)
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
				return nil, nil, err
			}

			_, err = io.Copy(wr, br)

			if err != nil {
				wr.Abort()
				return nil, nil, err
			}

			err = wr.Close()

			if err != nil {
				return nil, nil, err
			}

			logger.Info("Retrieved remote SQLite (%s) and stored as %s", remote_sqlite, local_sqlite)
		}
	*/

	if opts.SQLite {

		select {

		case <-ctx.Done():
			return nil, nil, nil
		default:
			// pass
		}

		if opts.LocalSQLite {

			f_opts := repo.DefaultFilenameOptions()
			fname := opts.Repo.SQLiteFilename(f_opts) // fmt.Sprintf("%s-latest.db", opts.Repo)

			local_sqlite = filepath.Join(opts.Workdir, fname)

		} else {

			d, err := database.BuildSQLite(ctx, local_checkout, opts)

			if err != nil {
				opts.Logger.Warning("Failed to build SQLlite %s because %s", local_sqlite, err)
				return nil, nil, err
			}

			// I don't necessarily believe this is being reported correctly but I
			// haven't been able to track down the errant reporting...
			// (20181127/thisisaaronland)

			opts.Logger.Status("Built %s without any reported errors", local_sqlite)

			distributions = append(distributions, d)
			local_sqlite = d.Path()
		}

		opts.Logger.Status("local sqlite is %s", local_sqlite)
	}

	_, err = os.Stat(local_sqlite)

	if err != nil {
		return nil, nil, err
	}

	// eventually we should be able to do these two operations in parallel
	// assuming that a SQLite database has been created - and the bundles
	// code has been updated to read from those databases...
	// (20180602/thisisaaronland)

	if opts.Meta {

		mode := "repo"
		source := local_checkout

		if opts.SQLite {
			mode = "sqlite"
			source = local_sqlite
		}

		opts.Logger.Status("build metafile from %s (%s)", mode, source)

		select {

		case <-ctx.Done():
			return nil, nil, nil
		default:
			// pass
		}

		ta := time.Now()

		d_many, err := csv.BuildMetaFiles(ctx, opts, mode, source)

		tb := time.Since(ta)

		if err != nil {
			return nil, nil, err
		}

		if len(d_many) == 0 {
			return nil, nil, err
		}

		for _, d := range d_many {
			distributions = append(distributions, d)
			local_metafiles = append(local_metafiles, d.Path())
		}

		opts.Logger.Status("time to build metafiles (%s) %v", strings.Join(local_metafiles, ","), tb)
	}

	if opts.Bundle {

		select {

		case <-ctx.Done():
			return nil, nil, nil
		default:
			// pass
		}

		ta := time.Now()

		// see notes in fs/bundles.go about meta/bundle filenames (20180731/thisisaaronland)

		bundle_dist, err := fs.BuildBundle(ctx, opts, local_metafiles, local_sqlite)

		tb := time.Since(ta)
		opts.Logger.Status("time to build bundles (%s) %v", strings.Join(local_bundlefiles, ","), tb)

		if err != nil {
			return nil, nil, err
		}

		if len(bundle_dist) == 0 {
			return nil, nil, errors.New("No metafiles produced")
		}

		for _, d := range bundle_dist {
			distributions = append(distributions, d)
			local_bundlefiles = append(local_bundlefiles, d.Path())
		}
	}

	meta := &dist.MetaData{
		CommitHash: commit_hash,
		Repo:       opts.Repo.Name(),
	}

	return distributions, meta, nil
}
